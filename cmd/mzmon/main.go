package main

import (
	"flag"
	"log"
	"time"

	"github.com/egymgmbh/dns-tools/gcp"
	influx "github.com/egymgmbh/dns-tools/influx"
	"github.com/egymgmbh/dns-tools/lib"
	metrics "github.com/rcrowley/go-metrics"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

func main() {
	gcpSAFile := flag.String("gcp-sa-file", "/secret/gcp-sa.json",
		"Google Cloud Platform Service Account file in JSON format.")
	influxConfigFile := flag.String("influx-config-file", "/secret/influx.json",
		"InfluxDB configuration file in JSON format.")
	pauseStr := flag.String("pause", "5m", "Pause between check runs.")
	flag.Parse()

	service, projectID, err := gcp.GetDNSService(*gcpSAFile, true)
	if err != nil {
		log.Fatalf("DNS API service: %v", err)
	}

	var iconf influx.Config
	err = iconf.Load(*influxConfigFile)
	if err != nil {
		log.Fatalf("load InfluxDB config: %v", err)
	}

	pause, err := time.ParseDuration(*pauseStr)
	if err != nil {
		log.Fatalf("invalid pause '%s': %v", *pauseStr, err)
	}

	// fire up influx client
	// pause/2: good old oversampling :)
	go influxdb.InfluxDB(metrics.DefaultRegistry, pause/2,
		iconf.Server, iconf.Database, iconf.Username, iconf.Password)

	// overall counters
	gaugeOK := metrics.NewGauge()
	metrics.Register("nsmon_"+projectID+"_total.OK", gaugeOK)
	gaugeMismatch := metrics.NewGauge()
	metrics.Register("nsmon_"+projectID+"_total.Mismatch", gaugeMismatch)
	gaugeError := metrics.NewGauge()
	metrics.Register("nsmon_"+projectID+"_total.Error", gaugeError)

	// main loop, where we check all managed zones and their delegations
	mzMetrics := make(map[string]metrics.Gauge)
	for {
		// fetch current managed zones
		managedZones, err := service.ManagedZones.List(projectID).Do()
		if err != nil {
			log.Fatalf("list managed zones: %v", err)
		}

		// sometimes, we discover new zones and need to create gauges in-the-fly
		for _, managedZone := range managedZones.ManagedZones {
			if _, ok := mzMetrics[managedZone.Name]; !ok {
				mzMetrics[managedZone.Name] = metrics.NewGauge()
				metrics.Register("nsmon_"+projectID+"_managedzone."+managedZone.Name,
					mzMetrics[managedZone.Name])
			}
		}

		statsError := int64(0)
		statsOK := int64(0)
		statsMismatch := int64(0)

		for _, managedZone := range managedZones.ManagedZones {
			curNameServers, err := lib.Lookup(managedZone.DnsName, "NS")
			if err != nil {
				mzMetrics[managedZone.Name].Update(-1)
				statsError++
				continue
			}
			if lib.RDatasEqual(curNameServers, managedZone.NameServers) &&
				len(curNameServers) > 0 {
				mzMetrics[managedZone.Name].Update(1)
				statsOK++
			} else {
				mzMetrics[managedZone.Name].Update(0)
				statsMismatch++
			}
		}

		gaugeError.Update(statsError)
		gaugeOK.Update(statsOK)
		gaugeMismatch.Update(statsMismatch)
		log.Printf("%v OK, %v Mismatch, %v Error", statsOK, statsMismatch, statsError)

		time.Sleep(pause)
	}
}
