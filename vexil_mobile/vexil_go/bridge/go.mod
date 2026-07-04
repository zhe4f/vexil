module vexil_go/bridge

go 1.25.9

require vexil_go v0.0.0

require (
	github.com/hashicorp/mdns v1.0.7 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	golang.org/x/mobile v0.0.0-20260611195102-4dd8f1dbf5d2 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/tools v0.46.0 // indirect
)

replace vexil_go => ../

tool golang.org/x/mobile/cmd/gobind
