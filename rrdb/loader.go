package rrdb

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"github.com/egymgmbh/dns-tools/lib"

	yaml "gopkg.in/yaml.v2"
)

// YAMLMailserver struct to load YAML data into: A singe mailserver with hostname
// and preference
type YAMLMailserver struct {
	Mailserver string
	Preference uint16
}

// YAMLMail struct to load YAML data into: A list of mailservers and an optional
// TTL
type YAMLMail struct {
	TTL         int
	Mailservers []YAMLMailserver
}

// YAMLForwarding struct to load YAML data into: A DNS forwarding via canonical
// name and optional TTL
type YAMLForwarding struct {
	TTL    int
	Target string
}

// YAMLDelegation struct to load YAML data into: A list of nameserver
// responsible for a delegated zone and an optional TTL
type YAMLDelegation struct {
	TTL         int
	Nameservers []string
}

// YAMLAddresses struct to load YAML data into: A list of IP and legacy IP
// addresses and an optional TTL
type YAMLAddresses struct {
	TTL      int
	Literals []string
}

// YAMLTexts struct to load YAML data into: A list of textual data associated
// with a name and an optional TTL
type YAMLTexts struct {
	TTL  int
	Data []string
}

// YAMLName struct to load YAML data into: A single name (label), may contain one
// forwarding or one list of delegations or a combination of mailservers, texts
// and addresses
type YAMLName struct {
	Name        string
	Description string
	Forwarding  YAMLForwarding
	Delegation  YAMLDelegation
	Mail        YAMLMail
	Texts       YAMLTexts
	Addresses   YAMLAddresses
}

// YAMLTemplate struct to load YAML data into: A template containing names and an
// optional description
type YAMLTemplate struct {
	Template    string
	Description string
	Templates   []string
	Names       []YAMLName
}

// YAMLZone struct to load YAML data into: A zone definition containing TTL,
// description, templates and names. All optional but should hold at least one.
type YAMLZone struct {
	Zone        string
	Description string
	TTL         int
	Templates   []string
	Names       []YAMLName
}

// YAMLFile holds the full YAML file data
type YAMLFile struct {
	Templates []YAMLTemplate
	Zones     []YAMLZone
}

func (db *RRDB) loadNS(fqdn string, delegation YAMLDelegation) error {
	rdatas := []string{}
	for _, rdata := range delegation.Nameservers {
		rdatas = append(rdatas, strings.TrimSpace(rdata))
	}
	if len(rdatas) == 0 {
		return nil
	}
	return db.SetNS(fqdn, delegation.TTL, rdatas)
}

func (db *RRDB) loadMX(fqdn string, mail YAMLMail) error {
	rdatas := []string{}
	for _, mailserver := range mail.Mailservers {
		rdata := fmt.Sprintf("%d %s", mailserver.Preference,
			strings.TrimSpace(mailserver.Mailserver))
		rdatas = append(rdatas, rdata)
	}
	if len(rdatas) == 0 {
		return nil
	}
	return db.SetMX(fqdn, mail.TTL, rdatas)
}

func (db *RRDB) loadTXT(fqdn string, texts YAMLTexts) error {
	var err error
	for _, rdata := range texts.Data {
		err = db.AddTXT(fqdn, texts.TTL, strings.TrimSpace(rdata))
		if err != nil {
			break
		}
	}
	return err
}

func (db *RRDB) loadCNAME(fqdn string, forwarding YAMLForwarding) error {

	rdata := strings.TrimSpace(forwarding.Target)
	if len(rdata) == 0 {
		return nil
	}
	return db.SetCNAME(fqdn, forwarding.TTL, rdata)
}

func (db *RRDB) loadAddresses(fqdn string, addresses YAMLAddresses) error {
	aRDatas := []string{}
	aaaaRDatas := []string{}
	for _, literal := range addresses.Literals {
		err := lib.IsValidIPv4(literal)
		if err == nil {
			aRDatas = append(aRDatas, literal)
			continue
		}
		err = lib.IsValidIPv6(literal)
		if err == nil {
			aaaaRDatas = append(aaaaRDatas, literal)
			continue
		}
		return fmt.Errorf("invalid address: %v", literal)
	}
	if len(aRDatas) > 0 {
		err := db.SetA(fqdn, addresses.TTL, aRDatas)
		if err != nil {
			return err
		}
	}
	if len(aaaaRDatas) > 0 {
		err := db.SetAAAA(fqdn, addresses.TTL, aaaaRDatas)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *RRDB) loadNames(names []YAMLName, zone YAMLZone) error {
	for _, name := range names {
		fqdn := lib.MakeFQDN(name.Name, zone.Zone)
		err := db.loadNS(fqdn, name.Delegation)
		if err != nil {
			return fmt.Errorf("zone %v: name %v: load delegations: %v",
				zone.Zone, name.Name, err)
		}
		err = db.loadMX(fqdn, name.Mail)
		if err != nil {
			return fmt.Errorf("zone %v: name %v: load mailservers: %v",
				zone.Zone, name.Name, err)
		}
		err = db.loadTXT(fqdn, name.Texts)
		if err != nil {
			return fmt.Errorf("zone %v: name %v: load texts: %v",
				zone.Zone, name.Name, err)
		}
		err = db.loadCNAME(fqdn, name.Forwarding)
		if err != nil {
			return fmt.Errorf("zone %v: name %v: load forwarding: %v",
				zone.Zone, name.Name, err)
		}
		err = db.loadAddresses(fqdn, name.Addresses)
		if err != nil {
			return fmt.Errorf("zone %v: name %v: load addresses: %v",
				zone.Zone, name.Name, err)
		}
	}
	return nil
}

// NewFromDirectory creates a new database from a directory of YAML-formatted
// zonedata files
func NewFromDirectory(directory string) (*RRDB, error) {
	fnames, err := filepath.Glob(path.Join(directory, "*.yml"))
	if err != nil {
		return nil, err
	}

	yamlFiles := make(map[string]YAMLFile)
	for _, fname := range fnames {
		data, ferr := ioutil.ReadFile(fname)
		if ferr != nil {
			return nil, ferr
		}
		yamlFile := YAMLFile{}
		ferr = yaml.UnmarshalStrict(data, &yamlFile)
		if ferr != nil {
			return nil, fmt.Errorf("file %v: %v", fname, ferr)
		}
		yamlFiles[fname] = yamlFile
	}

	// build templates map
	templates := make(map[string]YAMLTemplate)
	for fname, yamlFile := range yamlFiles {
		for _, template := range yamlFile.Templates {
			if _, seen := templates[template.Template]; seen {
				return nil, fmt.Errorf("file %v: template %v: duplicate",
					fname, template)
			}
			if len(template.Names) == 0 {
				return nil, fmt.Errorf("file %v: template %v: empty",
					fname, template)
			}
			templates[template.Template] = template // uh-oh, naming is fun :)
		}
	}

	db := New()
	for fname, yamlFile := range yamlFiles {
		for _, zone := range yamlFile.Zones {
			// templates
			for _, template := range zone.Templates {
				if _, ok := templates[template]; !ok {
					return nil, fmt.Errorf("file %v: zone %v: template %v: not found",
						fname, zone.Zone, template)
				}
				err = db.loadNames(templates[template].Names, zone)
				if err != nil {
					return nil, fmt.Errorf("file %v: template %v: %v",
						fname, template, err)
				}
			}
			// zone entries
			err = db.loadNames(zone.Names, zone)
			if err != nil {
				return nil, fmt.Errorf("file %v: %v", fname, err)
			}
		}
	}
	if len(db.root.nodes) == 0 {
		return nil, fmt.Errorf("empty database")
	}
	return db, nil
}
