package main

import (
	"github.com/ajstarks/svgo"
	"log"
	"math"
	"strings"
)

// Represents a possible transition between two state. Directional.
type Transition struct {
	Symbols []string
	From    *State
	To      *State
}

// Represents a state that we can draw on the graph. It keeps not if it's
// accepting or the start state, and also all transitions to or from the node.
// It also keep track of the point it was graphed at, if it has been graphed.
type State struct {
	Name        string
	Accepting   bool
	Start       bool
	GraphedAt   [2]int
	Graphed     bool
	Children    bool
	Transitions []*Transition
}

// The graph is a simple collection of states with a defined start state.
type Graph struct {
	States []*State
	Start  *State
}

// Parses the input line, adding it to the graph. It accepts two formats:
//
//  "accepting state"      - sets a state to be accepting
//  "state1 symbol state2" - creates a transition from state1 to state2 over
//                           the given symbol
//
func (g *Graph) Parse(line string) {
	// Handle accepting states
	if strings.HasPrefix(line, ACCEPTING_PREFIX) {
		names := strings.TrimPrefix(line, ACCEPTING_PREFIX)
		for _, state := range strings.Split(names, " ") {
			g.GetState(state).Accepting = true
			log.Printf("Set state of `%s` to accepting.", state)
		}
		return
	}

	// Otherwise we're defining a transition. Look up the nodes.
	parts := strings.Split(line, " ")
	start := g.GetState(parts[0])
	stop := g.GetState(parts[2])

	// Create and register the transition.
	transition := Transition{Symbols: []string{parts[1]}, From: start, To: stop}
	start.AddTransition(transition)
	if stop != start {
		stop.AddTransition(transition)
	}
	log.Printf("Added transition `%s`.", line)

	// If this was the first transition, the `start` must be our entry point.
	if g.Start == nil {
		g.Start = start
		start.Start = true
		log.Printf("Defined `%s` as the start state.", start.Name)
	}
}

// Gets a state from the graph by its name, creating and appending a
// state to the list if it does not already exist
func (g *Graph) GetState(name string) *State {
	for _, state := range g.States {
		if state.Name == name {
			return state
		}
	}

	state := &State{
		Name:        name,
		Accepting:   false,
		Start:       false,
		Transitions: []*Transition{},
	}
	g.States = append(g.States, state)
	log.Printf("Created state `%s`.", name)

	return state
}

// Adds a new transition to the state, grouping together to's/from's
// and updating symbols if they already exist.
func (s *State) AddTransition(transition Transition) {
	for _, t := range s.Transitions {
		if t.To == transition.To && t.From == transition.From {
			t.Symbols = append(t.Symbols, transition.Symbols...)
			return
		}
	}

	s.Transitions = append(s.Transitions, &transition)
}

// Plots the state at the given x, y coordinates.
func (s *State) PlotSelf(canvas *svg.SVG, x, y int) {
	// Draw out current node and label for the state
	canvas.Circle(x, y, NODE_BASE_RAD, NODE_BASE_STYLE)
	canvas.Text(x, y, s.Name, NODE_NAME_STYLE)

	// If we're at the start, draw the arrow down connecting to this node.
	if s.Start {
		base := float64(NODE_BASE_RAD)
		tx, ty := x, y-int(base*0.65)
		canvas.Line(x, y-ENTRY_SIZE-NODE_BASE_RAD, tx, ty, LINE_STYLE)
		drawArrow(canvas, tx, ty, math.Pi*1.5)
	}
	// If we're accepting, draw the accepting circle around this node.
	if s.Accepting {
		canvas.Circle(x, y, NODE_ACCEPTING_RAD, NODE_ACCEPTING_STYLE)
	}

	// Mark that we have graphed this node
	s.GraphedAt = [...]int{x, y}
	s.Graphed = true
}

// Plots all unplotted items which are connected to this one.
func (s *State) PlotChildren(canvas *svg.SVG) {
	if s.Children {
		return
	}

	x, y := s.GraphedAt[0], s.GraphedAt[1]
	s.Children = true

	// Take note of all transitions we still need to graph
	needToGraph := []*Transition{}
	for _, transition := range s.Transitions {
		if !transition.To.Graphed {
			needToGraph = append(needToGraph, transition)
		}
	}

	if len(needToGraph) > 0 {

		// We'll graph each new node in a 180 degree arc down from this.
		// Calculate the angle separation.
		delta := math.Pi / float64(len(needToGraph))
		current := math.Pi + (delta / 2)
		for _, Transition := range needToGraph {
			Transition.To.PlotSelf(
				canvas,
				x+int(math.Cos(current)*float64(NODE_SEPARATION+NODE_BASE_RAD*2)),
				y-int(math.Sin(current)*float64(NODE_SEPARATION+NODE_BASE_RAD*2)),
			)

			current += delta
		}
	}

	// Now draw all connecting lines that originate from this node. At this
	// point we know that everything we connect to has be graphed. Just
	// a bit more math, then we're done ;)
	for _, transition := range s.Transitions {
		log.Printf("plotting to %s", transition.To.Name)
		// Take care of plotting self-referential transitions,..
		if transition.To.Name == s.Name && transition.From.Name == s.Name {
			drawSelfReference(canvas, x, y)
			drawLabel(canvas, x, y-NODE_CIRCLE_RAD-NODE_BASE_RAD, transition.GetSymbolSet())
		} else if transition.To != s {
			// Pull the coordinates of the target
			tx, ty := transition.To.GraphedAt[0], transition.To.GraphedAt[1]
			// Calculate the directional "sway" in the line.
			sway := float64(LINE_SWAY_CONNECT)
			// Get the angle to the target
			angle := math.Atan2(float64(tx-x), float64(ty-y))
			// Easy calculation of the start and end points
			startx := x + int(math.Sin(angle-sway)*float64(NODE_BASE_RAD))
			starty := y + int(math.Cos(angle-sway)*float64(NODE_BASE_RAD))
			endx := tx + int(math.Sin(angle+math.Pi+sway)*float64(NODE_BASE_RAD))
			endy := ty + int(math.Cos(angle+math.Pi+sway)*float64(NODE_BASE_RAD))

			// Now the curvature point, which is 90 degrees off the
			// midpoint of the line.
			curvex := endx + (startx-endx)/2 + int(math.Cos(angle+math.Pi)*float64(LINE_SWAY_CURVE))
			curvey := endy + (starty-endy)/2 - int(math.Sin(angle+math.Pi)*float64(LINE_SWAY_CURVE))

			// Finally, draw the line and label it!
			drawConnect(canvas, startx, starty, curvex, curvey, endx, endy)
			drawLabel(canvas, curvex, curvey, transition.GetSymbolSet())

			// Allow the node to draw its children/connections as well.
			transition.To.PlotChildren(canvas)
		}
	}
}

// Recursively plots the point, and all connected elements.
func (s *State) Plot(canvas *svg.SVG, x, y int) {
	s.PlotSelf(canvas, x, y)
	s.PlotChildren(canvas)
}

// Returns the set of symbols for the transition state in set-ish notation.
func (s *Transition) GetSymbolSet() string {
	return "{" + strings.Join(s.Symbols, ", ") + "}"
}

// Draws a "self" reference. Sweeps an arc above the node centered
// at x, y.
func drawSelfReference(canvas *svg.SVG, x, y int) {
	angle := float64(NODE_CIRCLE_ANGLE) / 180 * math.Pi

	canvas.Arc(
		x+int(math.Cos(math.Pi/2+angle)*float64(NODE_BASE_RAD)),
		y-int(math.Sin(math.Pi/2+angle)*float64(NODE_BASE_RAD)),
		NODE_CIRCLE_RAD, NODE_CIRCLE_RAD,
		0, true, true,
		x+int(math.Cos(math.Pi/2-angle)*float64(NODE_BASE_RAD)),
		y-int(math.Sin(math.Pi/2-angle)*float64(NODE_BASE_RAD)),
		LINE_STYLE,
	)

	drawArrow(
		canvas,
		x+int(math.Cos(math.Pi/2+angle)*float64(NODE_BASE_RAD)),
		y-int(math.Sin(math.Pi/2+angle)*float64(NODE_BASE_RAD)),
		angle+math.Pi,
	)
}

// Draws a quadradic bezier curcle from s to e, adding an arrow on the end.
func drawConnect(canvas *svg.SVG, sx, sy, cx, cy, ex, ey int) {
	canvas.Qbez(sx, sy, cx, cy, ex, ey, LINE_STYLE)

	// Calculate the angle of the ending intersection
	angle := math.Atan2(float64(ey-cy), float64(ex-cx)) + math.Pi
	drawArrow(canvas, ex, ey, angle)
}

// Draws an arrow centered at x, y, rotated at the given angle
func drawArrow(canvas *svg.SVG, x, y int, angle float64) {
	canvas.Polygon(
		[]int{
			x,
			x + int(float64(ARROW_LENGTH)*math.Cos(angle-math.Pi/4)),
			x + int(float64(ARROW_LENGTH)*math.Cos(angle+math.Pi/4)),
		},
		[]int{
			y,
			y + int(float64(ARROW_LENGTH)*math.Sin(angle-math.Pi/4)),
			y + int(float64(ARROW_LENGTH)*math.Sin(angle+math.Pi/4)),
		},
		ARROW_STYLE,
	)
}

// Draws a label, for a line.
func drawLabel(canvas *svg.SVG, x, y int, text string) {
	canvas.Text(x, y, text, NODE_LABEL_STYLE)
}
