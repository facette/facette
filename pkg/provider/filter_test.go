package provider

import (
	"reflect"
	"sync"
	"testing"

	"github.com/facette/facette/pkg/catalog"
	"github.com/facette/facette/pkg/config"
)

func Test_Filter_Rewrite(test *testing.T) {
	expected := []catalog.Record{
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.octets.rx", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_octets.rx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.octets.tx", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_octets.tx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.packets.rx", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_packets.rx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "net.eth0.packets.tx", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "interface-eth0.if_packets.tx"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "load.load.shortterm", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "load.load.midterm", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1_example_net", Metric: "load.load.longterm", OriginalOrigin: "collectd",
			OriginalSource: "host1.example.net", OriginalMetric: "load.load.longterm"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.octets.rx", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_octets.rx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.octets.tx", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_octets.tx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.packets.rx", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_packets.rx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "net.eth0.packets.tx", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "interface-eth0.if_packets.tx"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "load.load.shortterm", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "load.load.midterm", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host2_example_net", Metric: "load.load.longterm", OriginalOrigin: "collectd",
			OriginalSource: "host2.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Action: "rewrite", Target: "source", Pattern: "\\.", Into: "_"},
		{Action: "rewrite", Target: "metric", Pattern: "^interface-(.+)\\.if_(.+)\\.(.+)$", Into: "net.$1.$2.$3"},
	}, len(expected))

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func Test_Filter_Discard(test *testing.T) {
	expected := []catalog.Record{
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host2.example.net", Metric: "load.load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host2.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Action: "discard", Target: "source", Pattern: "host1\\.example\\.net"},
		{Action: "discard", Target: "metric", Pattern: "^interface"},
	}, len(expected))

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func Test_Filter_Sieve(test *testing.T) {
	expected := []catalog.Record{
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Action: "sieve", Target: "source", Pattern: "host1\\.example\\.net"},
		{Action: "sieve", Target: "metric", Pattern: "load"},
	}, len(expected))

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func Test_Filter_Combined(test *testing.T) {
	expected := []catalog.Record{
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.shortterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.shortterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.midterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.midterm"},
		{Origin: "collectd", Source: "host1.example.net", Metric: "load.longterm",
			OriginalOrigin: "collectd", OriginalSource: "host1.example.net", OriginalMetric: "load.load.longterm"},
	}

	actual := runTestFilter([]*config.ProviderFilterConfig{
		{Action: "sieve", Target: "source", Pattern: "host1\\.example\\.net"},
		{Action: "discard", Target: "metric", Pattern: "interface"},
		{Action: "rewrite", Target: "metric", Pattern: "load\\.load", Into: "load"},
	}, len(expected))

	if !reflect.DeepEqual(expected, actual) {
		test.Logf("\nExpected %s\nbut got  %s", expected, actual)
		test.Fail()
	}
}

func runTestFilter(filters []*config.ProviderFilterConfig, outputCount int) []catalog.Record {
	var filteredRecords []catalog.Record

	testRecords := []catalog.Record{
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

	wg := &sync.WaitGroup{}
	wg.Add(outputCount)

	filterOutput := make(chan *catalog.Record)

	filterChain := newFilterChain(filters, filterOutput)

	go func(recordChan chan *catalog.Record, records *[]catalog.Record) {
		for {
			r := <-recordChan
			if r == nil {
				return
			}

			*records = append(*records, *r)
			wg.Done()
		}
	}(filterOutput, &filteredRecords)

	for i := range testRecords {
		filterChain.Input <- &testRecords[i]
	}

	wg.Wait()

	close(filterChain.Input)
	close(filterOutput)

	return filteredRecords
}
