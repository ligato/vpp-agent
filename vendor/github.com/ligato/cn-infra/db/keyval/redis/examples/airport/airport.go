//go:generate protoc --proto_path=./model --gogo_out=./model ./model/flight.proto
package main

import (
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"fmt"

	"sort"
	"strconv"

	"github.com/ligato/cn-infra/db"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/db/keyval/redis"
	"github.com/ligato/cn-infra/db/keyval/redis/examples/airport/model"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/cn-infra/utils/config"
	"math"
	"sync/atomic"
)

var diagram = `
                                                 =========
                                                  Airport
                                                 =========


         +---------+               +----------+
 put --->| Arrival |               |  Runway  |
         |---------|               |----------|
         |        --- put event -->|          |
         |         |<--- delete ---|          |                                        +-----------+
         +---------+               |  land    |                                        |  Hangar   |
                                   |          |< maintenance? > [yes] -- put w/ TTL -->|-----------|
                                   |          |     [no]                               | (expired) |
                                   |          |       |         +-----------+          +----|------+
                                   |          |       +- put -->| Departure |<-- put <-- del event
                                   |          |                 |-----------|
                                   |          |<--- put event ----          |
                                   |          |---- delete ---->|           |
                                   | take off |                 +-----------+
                                   +----------+

`

var log = logroot.Logger()

const (
	arrival            = "Arrival"
	departure          = "Departure"
	runway             = "Runway"
	hangar             = "Hangar"
	runwayLength       = 30
	runwayInterval     = 0.02
	runwayClearance    = 0.4
	runwaySpeedBump    = 4.0 / 9.0
	hangarThreshold    = 2.0 / 5.0
	hangarSlotCount    = 3
	hangarDurationLow  = 2
	hangarDurationHigh = 6
	flightIDLength     = 5
	flightSlotCount    = 5
	flightStatusSize   = 2*flightSlotCount + hangarSlotCount + 1
	flightIDFormat     = "%s%02d"
	hangarKeyFormat    = "%2s%2d:%d"
	columnSep          = "      "
	redisPause         = 0.05
)

var motions = []string{" ->", "<- "}

var flightStatusFormat = "\r"

var flightRadar = make(map[string]struct{})
var flightRadarMutex sync.Mutex

// priority of the flight
// For demo clarity, this is the order in which the flights arrive.  But its value
// can be set to represent other priority, such as fuel level.
var priority uint32

type priorities []uint32

var arrivalChan = make(chan keyval.ProtoWatchResp, flightSlotCount)
var departureChan = make(chan keyval.ProtoWatchResp, flightSlotCount)
var hangarChan = make(chan keyval.ProtoWatchResp, hangarSlotCount)
var runwayChan = make(chan flight.Info, flightSlotCount)

var redisConn *redis.BytesConnectionRedis
var arrivalBroker keyval.ProtoBroker
var arrivalWatcher keyval.ProtoWatcher
var departureBroker keyval.ProtoBroker
var departureWatcher keyval.ProtoWatcher
var hangarBroker keyval.ProtoBroker
var hangarWatcher keyval.ProtoWatcher

func main() {
	setup()
	startSimulation()
}

func setup() {
	rand.Seed(time.Now().UnixNano())

	printHeaders()

	setupFlightStatusFormat()

	redisConn = createConnection(os.Args[1])

	var arrivalProto, departureProto, hangarProto *kvproto.ProtoWrapper
	arrivalProto = kvproto.NewProtoWrapper(redisConn)
	departureProto = kvproto.NewProtoWrapper(redisConn)
	hangarProto = kvproto.NewProtoWrapper(redisConn)

	arrivalBroker = arrivalProto.NewBroker(arrival)
	arrivalWatcher = arrivalProto.NewWatcher(arrival)

	departureBroker = departureProto.NewBroker(departure)
	departureWatcher = departureProto.NewWatcher(departure)

	hangarBroker = hangarProto.NewBroker(hangar)
	hangarWatcher = hangarProto.NewWatcher(hangar)

	cleanup(false)

	arrivalWatcher.Watch(arrivalChan, "")
	departureWatcher.Watch(departureChan, "")
	hangarWatcher.Watch(hangarChan, "")
}

func createConnection(yamlFile string) *redis.BytesConnectionRedis {
	var err error
	var nodeClient redis.NodeClientConfig
	err = config.ParseConfigFromYamlFile(yamlFile, &nodeClient)
	if err != nil {
		log.Panicf("ParseConfigFromYamlFile() failed: %s", err)
	}
	pool, err := redis.CreateNodeClientConnPool(nodeClient)
	if err != nil {
		log.Panicf("CreateNodeClientConnPool() failed: %s", err)
	}
	var redisConn *redis.BytesConnectionRedis
	redisConn, err = redis.NewBytesConnectionRedis(pool, log)
	if err != nil {
		pool.Close()
		log.Panicf("NewBytesConnectionRedis() failed: %s", err)
	}
	return redisConn
}

func cleanup(report bool) {
	if report {
		fmt.Println("clean up")
		printFlightCounts()
	}
	arrivalBroker.Delete("")
	departureBroker.Delete("")
	hangarBroker.Delete("")
	if report {
		printFlightCounts()
	}
}

func startSimulation() {
	runArrivals()
	runDepartures()
	runHangar()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	for {
		select {
		case <-sigChan:
			fmt.Printf("\nReceived %v.\n", os.Interrupt)
			cleanup(true)
			redisConn.Close()
			os.Exit(1)
		case f, ok := <-runwayChan:
			if ok {
				processRunway(f)
				sleep(runwayClearance, runwayClearance)
			} else {
				log.Errorf("<-runwayChan returned false")
			}
		}
	}
}

func printHeaders() {
	fmt.Println()
	fmt.Println(diagram)
	fmt.Println()
	fmt.Println()

	diff := runwayLength - len(runway)
	pad := diff / 2
	var pad2 int
	if diff%2 == 0 {
		pad2 = pad
	} else {
		pad2 = pad + 1
	}
	fmt.Printf("%*s%s%*s%s%*s%s%-*s%s%s\n",
		flightIDLength*flightSlotCount, arrival,
		columnSep, pad, "", runway, pad2, "",
		columnSep, flightIDLength*flightSlotCount, departure,
		columnSep, hangar)
	dash60 := "-----------------------------------------------------------"
	waitingGuide := dash60[0 : flightIDLength*flightSlotCount]
	runwayGuide := dash60[0:runwayLength]
	hangarGuide := dash60[0 : flightIDLength*hangarSlotCount]
	fmt.Printf("%s%s%s%s%s%s%s\n",
		waitingGuide, columnSep, runwayGuide, columnSep, waitingGuide, columnSep, hangarGuide)
}

func printFlightCounts() {
	arrivals := countFlights(arrivalBroker, arrival)
	departures := countFlights(departureBroker, departure)
	hangars, err := getHangarFlights()
	if err != nil {
		log.Errorf("printFlightCounts() failed: %s", err)
	}
	fmt.Printf("arrivals %d, departures %d, hangars %d\n", arrivals, departures, len(hangars))
}

func runArrivals() {
	go func() {
		for i := 0; i < flightSlotCount-1; i++ {
			newArrival()
		}
		pause := 2*(runwayClearance+runwayInterval*float64(runwayLength-flightIDLength)) +
			(hangarDurationLow+hangarDurationHigh)*hangarThreshold/2 +
			9*redisPause
		low := pause - 0.3*pause
		high := pause + 0.3*pause
		for {
			newArrival()
			sleep(low, high)
		}
	}()

	go func() {
		for {
			r, ok := <-arrivalChan
			if ok {
				processArrival(r)
			} else {
				log.Errorf("<-arrivalChan returned false")
			}
		}
	}()
}

func runDepartures() {
	go func() {
		for {
			r, ok := <-departureChan
			if ok {
				processDeparture(r)
			} else {
				log.Errorf("<-departureChan returned false")
			}
		}
	}()
}

func runHangar() {
	go func() {
		for {
			r, ok := <-hangarChan
			if ok {
				processHangar(r)
			} else {
				log.Errorf("<-hangarChan returned false")
			}
		}
	}()
}

func setupFlightStatusFormat() {
	size := strconv.Itoa(flightIDLength)
	for i := 0; i < flightSlotCount; i++ {
		flightStatusFormat += "%" + size + "s"
	}
	flightStatusFormat += columnSep + "%-" + strconv.Itoa(runwayLength) + "s" + columnSep
	for i := 0; i < flightSlotCount; i++ {
		flightStatusFormat += "%-" + size + "s"
	}
	flightStatusFormat += columnSep
	for i := 0; i < hangarSlotCount; i++ {
		flightStatusFormat += "%-" + size + "s"
	}
}

func newArrival() {
	f := randomFlight()
	err := arrivalBroker.Put(flightID(f), &f)
	if err != nil {
		log.Errorf("newArrival() failed: %s", err)
	}
}

func randomFlight() flight.Info {
	airlines := []string{"AA", "DL", "SW", "UA"}
	numAirlines := len(airlines)

	atomic.AddUint32(&priority, 1)
	for {
		f := flight.Info{
			Airline:  airlines[rand.Int()%numAirlines],
			Number:   rand.Uint32()%99 + 1,
			Priority: priority,
		}
		var exists bool
		id := flightID(f)
		flightRadarMutex.Lock()
		if _, exists = flightRadar[id]; !exists {
			flightRadar[id] = struct{}{}
		}
		flightRadarMutex.Unlock()
		if !exists {
			return f
		}
	}
}

func flightID(flight flight.Info) string {
	return fmt.Sprintf(flightIDFormat, flight.Airline, flight.Number)
}

func processArrival(r keyval.ProtoWatchResp) {
	switch r.GetChangeType() {
	case db.Put:
		go func() {
			f := flight.Info{}
			r.GetValue(&f)
			f.Status = flight.Status_arrival
			runwayChan <- f
		}()
	case db.Delete:
		log.Debugf("%s deleted\n", r.GetKey())
	}
}

func processDeparture(r keyval.ProtoWatchResp) {
	switch r.GetChangeType() {
	case db.Put:
		go func() {
			f := flight.Info{}
			r.GetValue(&f)
			f.Status = flight.Status_departure
			runwayChan <- f
		}()
	case db.Delete:
		log.Debugf("%s deleted\n", r.GetKey())
	}
}

func processHangar(r keyval.ProtoWatchResp) {
	switch r.GetChangeType() {
	case db.Put:
		log.Debugf("%s updated\n", r.GetKey())
	case db.Delete:
		key := r.GetKey()
		f := flight.Info{}
		scanHangarKey(key, &f)
		err := departureBroker.Put(flightID(f), &f)
		if err != nil {
			log.Errorf("processHangar() failed: %s", err)
		}
	}
}

func makeHangarKey(f flight.Info) string {
	return fmt.Sprintf(hangarKeyFormat, f.Airline, f.Number, f.Priority)
}

func scanHangarKey(key string, f *flight.Info) error {
	_, err := fmt.Sscanf(key, hangarKeyFormat, &(f.Airline), &(f.Number), &(f.Priority))
	if err != nil {
		return err
	}
	return nil
}

func getHangarFlights() ([]flight.Info, error) {
	keys, err := hangarBroker.ListKeys("")
	if err != nil {
		return nil, fmt.Errorf("getHangarFlights() failed: %s", err)
	}

	flights := []flight.Info{}
	for {
		k, _, last := keys.GetNext()
		if last {
			break
		}
		f := flight.Info{}
		scanHangarKey(k, &f)
		flights = append(flights, f)
	}
	return flights, nil
}

func processRunway(f flight.Info) {
	id := flightID(f)

	if f.Status == flight.Status_arrival {
		log.Debugf("%s%s approaching runway\n", id, motions[f.Status])
		_, err := arrivalBroker.Delete(id)
		if err != nil {
			log.Errorf("processRunway(%s) failed: %s", id, err)
		}
		land(f)
		if rand.Float64() <= hangarThreshold {
			err = hangarBroker.Put(makeHangarKey(f), &f, keyval.WithTTL(randomDuration(hangarDurationLow, hangarDurationHigh)))
		} else {
			err = departureBroker.Put(id, &f)
		}
		if err != nil {
			log.Errorf("processRunway(%s) failed: %s", id, err)
		}
	} else {
		log.Debugf("%s%s approaching runway\n", id, motions[f.Status])
		_, err := departureBroker.Delete(id)
		if err != nil {
			log.Errorf("processRunway(%s) failed: %s", id, err)
		}
		takeOff(f)
		flightRadarMutex.Lock()
		delete(flightRadar, id)
		flightRadarMutex.Unlock()
	}
}

func land(f flight.Info) {
	flightInMotion := flightID(f) + motions[f.Status]
	size := len(flightInMotion)
	steps := runwayLength - size + 1
	interval := runwayInterval / 2
	var flightStatus = make([]interface{}, flightStatusSize)
	for i := 0; i < steps; i++ {
		flightStatus[flightSlotCount] = fmt.Sprintf("%*s", size+i, flightInMotion)
		fillArrivalStatus(flightStatus)
		fillDepartureStatus(flightStatus)
		fillHangarStatus(flightStatus)
		fmt.Printf(flightStatusFormat, flightStatus...)
		sleep(interval, interval)
		if i >= int(float64(steps)*(1-runwaySpeedBump)) {
			interval += runwayInterval
		}
	}
}

func takeOff(f flight.Info) {
	flightInMotion := motions[f.Status] + flightID(f)
	steps := runwayLength - len(flightInMotion) + 1
	interval := runwayInterval/2 + runwayInterval*math.Floor(float64(steps)*runwaySpeedBump)
	var flightStatus = make([]interface{}, flightStatusSize)
	for i := 0; i < steps; i++ {
		flightStatus[flightSlotCount] = fmt.Sprintf("%*s", runwayLength-i, flightInMotion)
		fillArrivalStatus(flightStatus)
		fillDepartureStatus(flightStatus)
		fillHangarStatus(flightStatus)
		fmt.Printf(flightStatusFormat, flightStatus...)
		sleep(interval, interval)
		if i < int(float64(steps)*runwaySpeedBump) {
			interval -= runwayInterval
		}
	}
}

func fillArrivalStatus(flightStatus []interface{}) {
	arrivals, err := getSortedFlights(arrivalBroker, arrival)
	if err != nil {
		log.Errorf("fillArrivalStatus() failed: %s", err)
		return
	}
	for i := 0; i < flightSlotCount; i++ {
		flightStatus[i] = ""
	}

	count := len(arrivals)
	if count > 0 {
		if count > flightSlotCount {
			count = flightSlotCount
		}
		for i := 0; i < count; i++ {
			flightStatus[flightSlotCount-1-i] = flightID(arrivals[i])
		}
	}
}

func fillDepartureStatus(flightStatus []interface{}) {
	departures, err := getSortedFlights(departureBroker, departure)
	if err != nil {
		log.Errorf("fillDepartureStatus() failed: %s", err)
		return
	}

	for i := flightSlotCount + 1; i < flightSlotCount*2+1; i++ {
		flightStatus[i] = ""
	}

	count := len(departures)
	if count > 0 {
		if count > flightSlotCount {
			count = flightSlotCount
		}
		for i := 0; i < count; i++ {
			flightStatus[flightSlotCount+1+i] = flightID(departures[i])
		}
	}
}

func fillHangarStatus(flightStatus []interface{}) {
	hangars, err := getHangarFlights()
	if err != nil {
		log.Errorf("fillHangarStatus() failed: %s", err)
		return
	}

	for i := flightSlotCount*2 + 1; i < flightStatusSize; i++ {
		flightStatus[i] = ""
	}

	count := len(hangars)
	if count > 0 {
		if count > hangarSlotCount {
			count = hangarSlotCount
		}
		for i := 0; i < count; i++ {
			flightStatus[flightSlotCount*2+1+i] = flightID(hangars[i])
		}
	}
}
func countFlights(broker keyval.ProtoBroker, name string) int {
	flights, err := getSortedFlights(broker, name)
	if err != nil {
		log.Errorf(err.Error())
		return 0
	}
	return len(flights)
}

func getSortedFlights(broker keyval.ProtoBroker, name string) ([]flight.Info, error) {
	kvi, err := broker.ListValues("")
	if err != nil {
		return nil, fmt.Errorf("getSortedFlights(%s) failed: %s", name, err)
	}
	var kvMap = make(map[uint32]flight.Info)
	var priorities = priorities{}
	for {
		kv, done := kvi.GetNext()
		if done {
			break
		}
		f := flight.Info{}
		kv.GetValue(&f)
		priorities = append(priorities, f.Priority)
		kvMap[f.Priority] = f
	}
	if len(priorities) == 0 {
		return []flight.Info{}, nil
	}
	sort.Sort(priorities)
	var flights = make([]flight.Info, len(priorities))
	for i, k := range priorities {
		flights[i] = kvMap[k]
	}
	return flights, nil
}

func getPrefix(broker keyval.BytesBroker) string {
	if b, yes := broker.(*redis.BytesBrokerWatcherRedis); yes {
		return b.GetPrefix()
	}
	return ""
}

func sleep(lowSeconds float64, highSeconds float64) {
	time.Sleep(randomDuration(lowSeconds, highSeconds))
}

func randomDuration(lowSeconds float64, highSeconds float64) time.Duration {
	nanos := lowSeconds * 1e9
	if highSeconds != lowSeconds {
		nanos += (highSeconds - lowSeconds) * 1e9 * rand.Float64()
	}
	return time.Duration(int64(nanos))
}

func (p priorities) Len() int           { return len(p) }
func (p priorities) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p priorities) Less(i, j int) bool { return p[i] < p[j] }
