module github.com/gverger/optimview

go 1.23.0

require (
	github.com/gen2brain/raylib-go/raygui v0.0.0-20241117153000-01864c04b849
	github.com/gen2brain/raylib-go/raylib v0.0.0-20241117153000-01864c04b849
	github.com/nikolaydubina/go-graph-layout v0.0.0-20240509045315-dafeb51fdd74
	github.com/phuslu/log v1.0.113
)

require (
	github.com/ebitengine/purego v0.8.1 // indirect
	golang.org/x/exp v0.0.0-20241108190413-2d47ceb2692f // indirect
	golang.org/x/sys v0.27.0 // indirect
	gonum.org/v1/gonum v0.9.3 // indirect
)

replace github.com/nikolaydubina/go-graph-layout v0.0.0-20240509045315-dafeb51fdd74 => ../go-graph-layout
