SOURCES := /Users/shreyaskali/Startups/Me/Deferred MBA/University Planner
INTERSECTIONS := /Users/shreyaskali/Developer/AI x Futurist/intersections

.PHONY: build test gen doctor tui clean

build:
	go build -o uplanner ./cmd/uplanner

test:
	go test ./...

gen:
	go run ./cmd/gen-data --uplanner "$(SOURCES)" --intersections "$(INTERSECTIONS)" --out ./content

doctor: build
	./uplanner doctor

tui: build
	./uplanner tui

clean:
	rm -f uplanner
