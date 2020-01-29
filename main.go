package main

import (
	crand "crypto/rand"
	"flag"
	"math"
	"math/big"
	"math/rand"
	"os"
	"runtime"
	"time"
)

type simNum struct {
	x         float64
	y         float64
	direction float64
	speed     float64
	viewAngle float64
	viewR	  float64
}

func main() {
	seed, _ := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	rand.Seed(seed.Int64())
	var (
		chooseRight = rand.Float64()
		wall        float64
		timerSec	int
		agents		uint
	)
	flag.Float64Var(&wall, "wall", 100.0, "limit x-max, y-max(min = 0, default = 100.0)")
	flag.IntVar(&timerSec, "time", 20, "exec time length before exit by sec (default = 20)")
	flag.UintVar(&agents, "agents", 120, "number of agents(default = 120)")
	flag.Parse()

	mtVars := make(chan int, agents)
	mtSync := make(chan int, agents)
	currentSlice := make ([]simNum, agents);
	for i := 0; i < len(currentSlice); i++ {
		currentX := rand.Float64() * wall
		currentY := rand.Float64() * wall
		currentSpeed := 1.0 + (rand.Float64() * 4.0)
		currentViewAngle := (math.Pi / 6.0) + (rand.Float64() * math.Pi * 5.0 / 6.0)
		currentViewR := rand.Float64() * 10.0
		currentDirection := rand.Float64() * 2 * math.Pi
		currentSlice[i] = simNum{x: currentX, y: currentY, direction: currentDirection, speed: currentSpeed, viewAngle: currentViewAngle, viewR: currentViewR}
	}

	futureSlice := make([]simNum, agents)

	nextSlice:= make([]simNum, agents)

	doneLoop := 0
	go func () {
		timer1 := time.NewTimer(time.Duration(timerSec) * time.Second)
		<- timer1.C
		println("sps = ", float64(doneLoop) / float64(timerSec))
		os.Exit(0)
	}()

	for loop := 0; ; loop++ {
		doneLoop = loop
		//println("simulation step:", loop)
		for i := 0; i < len(futureSlice); i++{
			mtVars <- i
		}
		for i := 0; i < len(futureSlice); i++ {
			go func() {
				j := <- mtVars
				futureX := currentSlice[j].x + currentSlice[j].speed * math.Cos(currentSlice[j].direction)
				futureY := currentSlice[j].y + currentSlice[j].speed * math.Sin(currentSlice[j].direction)
				futureDirection := currentSlice[j].direction
				for k := 0 ;; k++ {
					futureX, futureDirection, _ = boundCheck(futureX, wall, futureDirection, true, 0)
					if futureX == math.Mod(futureX, wall) {
						break
					}
				}
				for k:= 0 ;; k++ {
					futureY, futureDirection, _ = boundCheck(futureY, wall, futureDirection, false, 0)
					if futureY == math.Mod(futureY, wall) {
						break
					}
				}
				futureSlice[j] = simNum{x: futureX, y: futureY, direction: futureDirection, speed: currentSlice[j].speed, viewAngle: currentSlice[j].viewAngle, viewR: currentSlice[j].viewR}
				mtSync <- j
			}()
		}
		for i := 0; i < len(currentSlice); i++ {
			<- mtSync
			mtVars <- i
		}
		for i := 0; i < len(currentSlice); i++ {
			go func() {
				j := <- mtVars
				shortest := -1
				shortestR := math.MaxFloat64
				for k := 0; k < len(currentSlice); k++ {
					if j == k {
						continue
					}
					diffX := currentSlice[k].x - currentSlice[k].x
					diffY := currentSlice[k].y - currentSlice[k].y
					r := math.Pow(diffX, 2) + math.Pow(diffY, 2)
					if r <= math.Pow(currentSlice[j].viewR, 2) {
						angle := math.Atan(diffY / diffX)
						if (currentSlice[j].direction - currentSlice[j].viewAngle/2) <= angle {
							if angle <= (currentSlice[j].direction + currentSlice[j].viewAngle/2) {
								diffX = futureSlice[k].x - futureSlice[j].x
								diffY = futureSlice[k].y - futureSlice[j].y
								r = math.Pow(diffX, 2) + math.Pow(diffY, 2)
								if r < shortestR {
									shortestR = r
									shortest = k
								}
							}
						}
					}
				}
				if shortest == -1 {
					nextSlice[j] = futureSlice[j]
				} else {
					if chooseRight > rand.Float64() {
						nextSlice[j].direction = currentSlice[j].direction - math.Pi*rand.Float64()/3.0
					} else {
						nextSlice[j].direction = currentSlice[j].direction + math.Pi*rand.Float64()/3.0
					}
					nextSlice[j].x = currentSlice[j].x + currentSlice[j].speed * math.Cos(futureSlice[j].direction)
					nextSlice[j].y = currentSlice[j].y + currentSlice[j].speed * math.Sin(futureSlice[j].direction)
					for k := 0 ;; k++ {
						nextSlice[j].x, nextSlice[j].direction, _ = boundCheck(nextSlice[j].x, wall, nextSlice[j].direction, true, 0)
						if nextSlice[j].x == math.Mod(nextSlice[j].x, wall) {
							break
						}
					}
					for k := 0 ;; k++ {
						nextSlice[j].y, nextSlice[j].direction, _ = boundCheck(nextSlice[j].y, wall, nextSlice[j].direction, false, 0)
						if nextSlice[j].y == math.Mod(nextSlice[j].y, wall) {
							break
						}
					}
				}
				mtSync <- j
			}()
		}
		for i := 0; i < len(currentSlice); i++ {
			<- mtSync
		}
		copy(currentSlice, nextSlice)
	}
	runtime.Goexit()
}

func boundCheck(loc float64, wall float64, direction float64, isX bool, bound int) (float64, float64, int) {
	if loc > wall {
		loc = loc - (2 * (loc - wall))
		bound = bound + 1
	}
	if loc < 0 {
		loc = -loc
		bound = bound + 1
	}

	if bound%2 == 1 {
		if isX == true {
			if math.Mod(direction, math.Pi) < math.Pi/2 {
				direction = direction + math.Pi/2
			} else {
				direction = direction - math.Pi/2
			}
		} else {
			if math.Mod(direction, math.Pi) > math.Pi/2 {
				direction = direction + math.Pi/2
			} else {
				direction = direction - math.Pi/2
			}
		}
	}
	direction = math.Mod(direction, 2*math.Pi)
	return loc, direction, bound
}
