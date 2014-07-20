package provider

import (
	"reflect"
	"testing"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
)

func Test_Filter_Rewrite(test *testing.T) {
	expected := []catalog.CatalogRecord{
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.octets.rx",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_octets.rx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.octets.tx",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_octets.tx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.packets.rx",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_packets.rx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.packets.tx",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_packets.tx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "load.load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "load.load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "load.load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.longterm"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.octets.rx",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_octets.rx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.octets.tx",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_octets.tx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.packets.rx",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_packets.rx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.packets.tx",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_packets.tx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "load.load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "load.load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "load.load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Target: "source", Pattern: "\\.", Rewrite: "_"},
		{Target: "metric", Pattern: "^interface-(.+)\\.if_(.+)\\.(.+)$", Rewrite: "net.$1.$2.$3"},
	})

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func Test_Filter_Discard(test *testing.T) {
	expected := []catalog.CatalogRecord{
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Target: "source", Pattern: "host1\\.example\\.net", Discard: true},
		{Target: "metric", Pattern: "^interface", Discard: true},
	})

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func Test_Filter_Sieve(test *testing.T) {
	expected := []catalog.CatalogRecord{
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Target: "source", Pattern: "host1\\.example\\.net", Sieve: true},
		{Target: "metric", Pattern: "load", Sieve: true},
	})

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func Test_Filter_Combined(test *testing.T) {
	expected := []catalog.CatalogRecord{
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Target: "source", Pattern: "host1\\.example\\.net", Sieve: true},
		{Target: "metric", Pattern: "interface", Discard: true},
		{Target: "metric", Pattern: "load\\.load", Rewrite: "load"},
	})

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func runTestFilter(filters []*config.ProviderFilterConfig) []catalog.CatalogRecord {
	var testRecords = []catalog.CatalogRecord{
		{Origin: "collectd", Source: "host1.example.net", Metric: "interface-eth0.if_octets.rx"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "interface-eth0.if_octets.tx"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "interface-eth0.if_packets.rx"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "interface-eth0.if_packets.tx"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.longterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "interface-eth0.if_octets.rx"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "interface-eth0.if_octets.tx"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "interface-eth0.if_packets.rx"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "interface-eth0.if_packets.tx"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.midterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.longterm"},
	}

	filterOutput := make(chan *catalog.CatalogRecord)

	filterChain := newFilterChain(filters, filterOutput)

	filteredRecords := make([]catalog.CatalogRecord, 0)

	done := make(chan struct{})
	go func(doneChan chan struct{}, recordChan chan *catalog.CatalogRecord, records *[]catalog.CatalogRecord) {
		for {
			select {
			case <-doneChan:
				return
			case record := <-recordChan:
				*records = append(*records, *record)
			}
		}
	}(done, filterOutput, &filteredRecords)

	for i := range testRecords {
		filterChain.Input <- &testRecords[i]
	}

	done <- struct{}{}

	close(filterChain.Input)
	close(filterOutput)
	close(done)

	return filteredRecords
}
