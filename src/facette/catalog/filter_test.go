package catalog

import (
	"reflect"
	"sync"
	"testing"

	"facette/backend"
)

func Test_Filter_Rewrite(t *testing.T) {
	expected := []Record{
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.octets.rx", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "interface-eth0.if_octets.rx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.octets.tx", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "interface-eth0.if_octets.tx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.packets.rx", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "interface-eth0.if_packets.rx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.packets.tx", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "interface-eth0.if_packets.tx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "load.load.shortterm", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "load.load.shortterm"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "load.load.midterm", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "load.load.midterm"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "load.load.longterm", OriginalOrigin: "origin1",
			OriginalSource: "host1_example_net", OriginalMetric: "load.load.longterm"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.octets.rx", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "interface-eth0.if_octets.rx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.octets.tx", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "interface-eth0.if_octets.tx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.packets.rx", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "interface-eth0.if_packets.rx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.packets.tx", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "interface-eth0.if_packets.tx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "load.load.shortterm", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "load.load.shortterm"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "load.load.midterm", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "load.load.midterm"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "load.load.longterm", OriginalOrigin: "origin1",
			OriginalSource: "host2_example_net", OriginalMetric: "load.load.longterm"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-idle", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-idle"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-interrupt", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-interrupt"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-nice", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-nice"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-softirq", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-softirq"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-steal", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-steal"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-system", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-system"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-user", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-user"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-wait", OriginalOrigin: "origin2",
			OriginalSource: "host3_example_net", OriginalMetric: "cpu.percent-wait"},
	}

	result := runTestFilter(&backend.ProviderFilters{
		{Action: "rewrite", Target: "origin", Pattern: "^origin(\\d+)$", Into: "origin-$1"},
		{Action: "rewrite", Target: "source", Pattern: "_", Into: "."},
		{Action: "rewrite", Target: "metric", Pattern: "^interface-(.+)\\.if_(.+)\\.(.+)$", Into: "net.$1.$2.$3"},
	}, len(expected))

	if !reflect.DeepEqual(expected, result) {
		t.Logf("\nExpected %#v\nbut got  %#v", expected, result)
		t.Fail()
	}
}

func Test_Filter_Discard(t *testing.T) {
	expected := []Record{
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.shortterm",
			OriginalOrigin: "origin1", OriginalSource: "host2_example_net", OriginalMetric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.midterm",
			OriginalOrigin: "origin1", OriginalSource: "host2_example_net", OriginalMetric: "load.load.midterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.longterm",
			OriginalOrigin: "origin1", OriginalSource: "host2_example_net", OriginalMetric: "load.load.longterm"},
	}

	result := runTestFilter(&backend.ProviderFilters{
		{Action: "discard", Target: "origin", Pattern: "origin2"},
		{Action: "discard", Target: "source", Pattern: "host1_example_net"},
		{Action: "discard", Target: "metric", Pattern: "^interface"},
	}, len(expected))

	if !reflect.DeepEqual(expected, result) {
		t.Logf("\nExpected %#v\nbut got  %#v", expected, result)
		t.Fail()
	}
}

func Test_Filter_Sieve(t *testing.T) {
	expected := []Record{
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.shortterm",
			OriginalOrigin: "origin1", OriginalSource: "host1_example_net", OriginalMetric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.midterm",
			OriginalOrigin: "origin1", OriginalSource: "host1_example_net", OriginalMetric: "load.load.midterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.longterm",
			OriginalOrigin: "origin1", OriginalSource: "host1_example_net", OriginalMetric: "load.load.longterm"},
	}

	result := runTestFilter(&backend.ProviderFilters{
		{Action: "sieve", Target: "origin", Pattern: "origin1"},
		{Action: "sieve", Target: "source", Pattern: "host1_example_net"},
		{Action: "sieve", Target: "metric", Pattern: "load"},
	}, len(expected))

	if !reflect.DeepEqual(expected, result) {
		t.Logf("\nExpected %#v\nbut got  %#v", expected, result)
		t.Fail()
	}
}

func Test_Filter_Combined(t *testing.T) {
	expected := []Record{
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.shortterm",
			OriginalOrigin: "origin1", OriginalSource: "host1_example_net", OriginalMetric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.midterm",
			OriginalOrigin: "origin1", OriginalSource: "host1_example_net", OriginalMetric: "load.load.midterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.longterm",
			OriginalOrigin: "origin1", OriginalSource: "host1_example_net", OriginalMetric: "load.load.longterm"},
	}

	result := runTestFilter(&backend.ProviderFilters{
		{Action: "sieve", Target: "source", Pattern: "host1_example_net"},
		{Action: "discard", Target: "metric", Pattern: "interface"},
		{Action: "rewrite", Target: "metric", Pattern: "load\\.load", Into: "load"},
	}, len(expected))

	if !reflect.DeepEqual(expected, result) {
		t.Logf("\nExpected %#v\nbut got  %#v", expected, result)
		t.Fail()
	}
}

func runTestFilter(filters *backend.ProviderFilters, expectedLen int) []Record {
	testRecords := []Record{
		{Origin: "origin1", Source: "host1_example_net", Metric: "interface-eth0.if_octets.rx"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "interface-eth0.if_octets.tx"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "interface-eth0.if_packets.rx"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "interface-eth0.if_packets.tx"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.midterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.longterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "interface-eth0.if_octets.rx"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "interface-eth0.if_octets.tx"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "interface-eth0.if_packets.rx"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "interface-eth0.if_packets.tx"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.midterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.longterm"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-idle"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-interrupt"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-nice"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-softirq"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-steal"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-system"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-user"},
		{Origin: "origin2", Source: "host3_example_net", Metric: "cpu.percent-wait"},
	}

	wg := &sync.WaitGroup{}
	wg.Add(expectedLen)

	result := []Record{}
	chain := NewFilterChain(filters)

	go func(chain *FilterChain, records *[]Record) {
		for {
			select {
			case r := <-chain.Output:
				if r == nil {
					return
				}

				*records = append(*records, *r)
				wg.Done()

			case _ = <-chain.Messages:
				// consume messages to avoid test being blocked
			}
		}
	}(chain, &result)

	for i := range testRecords {
		chain.Input <- &testRecords[i]
	}

	wg.Wait()

	close(chain.Input)
	close(chain.Output)

	return result
}
