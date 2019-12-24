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
		timerSec	uint64
		agents		uint64
	)
	flag.Float64Var(&wall, "wall", 100.0, "limit x-max, y-max(min = 0, default = 100.0)")
	flag.Uint64Var(&timerSec, "time", 20, "exec time length before exit by sec (default = 50)")
	flag.Uint64Var(&agents, "agents", 60, "number of agents(default = 30)")
	flag.Parse()

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

	go exitTimer(timerSec)
	for loop := 0; ; loop++ {
		for i := 0; i < len(futureSlice); i++ {
			futureX := currentSlice[i].x + currentSlice[i].speed * math.Cos(currentSlice[i].direction)
			futureY := currentSlice[i].y + currentSlice[i].speed * math.Sin(currentSlice[i].direction)
			futureDirection := currentSlice[i].direction
			for loop := 0 ;; loop++ {
				futureX, futureDirection, _ = boundCheck(futureX, wall, futureDirection, true, 0)
				if futureX == math.Mod(futureX, wall) {
					break
				}
			}
			for loop := 0 ;; loop++ {
				futureY, futureDirection, _ = boundCheck(futureY, wall, futureDirection, false, 0)
				if futureY == math.Mod(futureY, wall) {
					break
				}
			}
			futureSlice[i] = simNum{x: futureX, y: futureY, direction: futureDirection, speed: currentSlice[i].speed, viewAngle: currentSlice[i].viewAngle, viewR: currentSlice[i].viewR}
		}
		print("simulation step:", loop)
		for i := 0; i < len(currentSlice); i++ {
			shortest := -1
			shortestR := math.MaxFloat64
			for j := 0; j < len(currentSlice); j++ {
				if i == j {
					continue
				}
				diffX := currentSlice[j].x - currentSlice[i].x
				diffY := currentSlice[j].y - currentSlice[i].y
				r := math.Pow(diffX, 2) + math.Pow(diffY, 2)
				if r <= math.Pow(currentSlice[i].viewR, 2) {
					angle := math.Atan(diffY / diffX)
					if (currentSlice[i].direction - currentSlice[i].viewAngle/2) <= angle {
						if angle <= (currentSlice[i].direction + currentSlice[i].viewAngle/2) {
							diffX = futureSlice[j].x - futureSlice[i].x
							diffY = futureSlice[j].y - futureSlice[i].y
							r = math.Pow(diffX, 2) + math.Pow(diffY, 2)
							if r < shortestR {
								shortestR = r
								shortest = j
							}
						}
					}
				}
			}
			if shortest == -1 {
				nextSlice[i] = futureSlice[i]
			} else {
				if chooseRight > rand.Float64() {
					nextSlice[i].direction = currentSlice[i].direction - math.Pi*rand.Float64()/3.0
				} else {
					nextSlice[i].direction = currentSlice[i].direction + math.Pi*rand.Float64()/3.0
				}
				nextSlice[i].x = currentSlice[i].x + currentSlice[i].speed * math.Cos(futureSlice[i].direction)
				nextSlice[i].y = currentSlice[i].y + currentSlice[i].speed * math.Sin(futureSlice[i].direction)
				for loop := 0 ;; loop++ {
					nextSlice[i].x, nextSlice[i].direction, _ = boundCheck(nextSlice[i].x, wall, nextSlice[i].direction, true, 0)
					if nextSlice[i].x == math.Mod(nextSlice[i].x, wall) {
						break
					}
				}
				for loop := 0 ;; loop++ {
					nextSlice[i].y, nextSlice[i].direction, _ = boundCheck(nextSlice[i].y, wall, nextSlice[i].direction, false, 0)
					if nextSlice[i].y == math.Mod(nextSlice[i].y, wall) {
						break
					}
				}
			}

		}
		copy(currentSlice, nextSlice)
	}
	runtime.Goexit()
}

func exitTimer(timeSec uint64) {
	timer1 := time.NewTimer(time.Duration(timeSec) * time.Second)
	<- timer1.C
	os.Exit(0)
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
