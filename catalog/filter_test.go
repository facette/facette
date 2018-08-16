package catalog

import (
	"sync"
	"testing"

	"facette.io/facette/storage"
	"github.com/stretchr/testify/assert"
)

func Test_Filter_Rewrite(t *testing.T) {
	expected := []Record{
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.octets.rx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.octets.tx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.packets.rx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "net.eth0.packets.tx"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "load.load.shortterm"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "load.load.midterm"},
		{Origin: "origin-1", Source: "host1.example.net", Metric: "load.load.longterm"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.octets.rx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.octets.tx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.packets.rx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "net.eth0.packets.tx"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "load.load.shortterm"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "load.load.midterm"},
		{Origin: "origin-1", Source: "host2.example.net", Metric: "load.load.longterm"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-idle"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-interrupt"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-nice"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-softirq"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-steal"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-system"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-user"},
		{Origin: "origin-2", Source: "host3.example.net", Metric: "cpu.percent-wait"},
	}

	assert.Equal(t, expected, runTestFilter(&storage.ProviderFilters{
		{Action: "rewrite", Target: "origin", Pattern: "^origin(\\d+)$", Into: "origin-$1"},
		{Action: "rewrite", Target: "source", Pattern: "_", Into: "."},
		{Action: "rewrite", Target: "metric", Pattern: "^interface-(.+)\\.if_(.+)\\.(.+)$", Into: "net.$1.$2.$3"},
	}, len(expected)))
}

func Test_Filter_Discard(t *testing.T) {
	expected := []Record{
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.midterm"},
		{Origin: "origin1", Source: "host2_example_net", Metric: "load.load.longterm"},
	}

	assert.Equal(t, expected, runTestFilter(&storage.ProviderFilters{
		{Action: "discard", Target: "origin", Pattern: "origin2"},
		{Action: "discard", Target: "source", Pattern: "host1_example_net"},
		{Action: "discard", Target: "metric", Pattern: "^interface"},
	}, len(expected)))
}

func Test_Filter_Sieve(t *testing.T) {
	expected := []Record{
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.shortterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.midterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.load.longterm"},
	}

	assert.Equal(t, expected, runTestFilter(&storage.ProviderFilters{
		{Action: "sieve", Target: "origin", Pattern: "origin1"},
		{Action: "sieve", Target: "source", Pattern: "host1_example_net"},
		{Action: "sieve", Target: "metric", Pattern: "load"},
	}, len(expected)))
}

func Test_Filter_Combined(t *testing.T) {
	expected := []Record{
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.shortterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.midterm"},
		{Origin: "origin1", Source: "host1_example_net", Metric: "load.longterm"},
	}

	assert.Equal(t, expected, runTestFilter(&storage.ProviderFilters{
		{Action: "sieve", Target: "source", Pattern: "host1_example_net"},
		{Action: "discard", Target: "metric", Pattern: "interface"},
		{Action: "rewrite", Target: "metric", Pattern: "load\\.load", Into: "load"},
	}, len(expected)))
}

func runTestFilter(filters *storage.ProviderFilters, expectedLen int) []Record {
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

			case <-chain.Messages:
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
