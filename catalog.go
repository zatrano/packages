package packages

import "github.com/zatrano/framework/v2/kernel"

// Catalog lists opt-in toolkit libraries. The consumer-facing name list is
// console/catalog.go in the framework; this file is the packages-module copy
// for toolkit/* (LayerAddon, KindLibrary).
var Catalog = []kernel.PackageInfo{
	{Name: "ai", Layer: kernel.LayerIntelligence, Kind: kernel.KindService, Stability: "experimental", Description: "AI chat providers"},
	{Name: "rag", Layer: kernel.LayerIntelligence, Kind: kernel.KindLibrary, Stability: "experimental", Description: "RAG chunking, embed pipeline, vector store helpers"},
	{Name: "agent", Layer: kernel.LayerIntelligence, Kind: kernel.KindLibrary, Stability: "experimental", Description: "AI agent loop, tools, conversation memory"},
	{Name: "toolkit/arr", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Array/slice helpers"},
	{Name: "toolkit/color", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Color helpers"},
	{Name: "toolkit/date", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Date/time helpers"},
	{Name: "toolkit/html", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "HTML helpers"},
	{Name: "toolkit/money", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Money helpers"},
	{Name: "toolkit/num", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Number helpers"},
	{Name: "toolkit/str", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "String helpers"},
}
