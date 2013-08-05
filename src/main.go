package main

import ("flag"
	"strings"
	"benchmarkHelpers"
)

const (
	persistentExchangeName = "bmexch"
	persistentQueueName    = "bmqueue"
)

func main() {
	urlsOpt := flag.String("urls", "amqp://guest:guest@localhost:5672/,amqp://guest:guest@localhost:5673/", "URLS to run the tests against")
	urls := strings.Split(*urlsOpt, ",")
	iterations := flag.Int("iterations", 1000, "Number of iterations")

	benchmarkPublishSubscribe := flag.Bool("runpubsub", true, "Benchmark publish/subscribe")
	benchmarkTransientQueueCreationCleanup := flag.Bool("runtransientqueue", true, "Benchmark transient queue creation/cleanup")
	benchmarkTemporaryQueueCreationCleanup := flag.Bool("runtemporaryqueue", true, "Benchmark temporary queue creation/cleanup")

	flag.Parse()

	for _, url := range urls {
		parm := benchmarkHelpers.BenchmarkParams{Url: url, Iterations: *iterations}
		if (*benchmarkPublishSubscribe) {benchmarkHelpers.BenchmarkPubSub(parm, persistentExchangeName, persistentQueueName)}
		if (*benchmarkTransientQueueCreationCleanup) {benchmarkHelpers.BenchmarkTransientQueueExchangeCreationDeletion(parm)}
		if (*benchmarkTemporaryQueueCreationCleanup) {benchmarkHelpers.BenchmarkTemporaryQueueExchangeCreationDeletion(parm)}
	}
}
