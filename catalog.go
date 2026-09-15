package packages

import "github.com/zatrano/framework/v2/kernel"

// Catalog lists experimental intelligence libraries and opt-in toolkit libraries.
// The consumer-facing name list is console/describe/catalog.go in the framework;
// this file is the packages-module copy.
var Catalog = []kernel.PackageInfo{
	{Name: "ai", Layer: kernel.LayerIntelligence, Kind: kernel.KindService, Stability: "experimental", Description: "AI chat providers"},
	{Name: "rag", Layer: kernel.LayerIntelligence, Kind: kernel.KindLibrary, Stability: "experimental", Description: "RAG chunking, embed pipeline, vector store helpers"},
	{Name: "agent", Layer: kernel.LayerIntelligence, Kind: kernel.KindLibrary, Stability: "experimental", Description: "AI agent loop, tools, conversation memory"},
	{Name: "workflow", Layer: kernel.LayerIntelligence, Kind: kernel.KindLibrary, Stability: "experimental", Description: "Generic process graphs (agents enter via agent.AsExecutor)"},
	{Name: "toolkit/arr", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Array/slice helpers"},
	{Name: "toolkit/bloom", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Bloom filter"},
	{Name: "toolkit/circuit", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Circuit breaker"},
	{Name: "toolkit/collection", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Collection helpers"},
	{Name: "toolkit/color", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Color helpers"},
	{Name: "toolkit/concurrency", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Concurrency primitives"},
	{Name: "toolkit/cron", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Cron expression parser"},
	{Name: "toolkit/date", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Date/time helpers"},
	{Name: "toolkit/debug", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Debug dump helpers"},
	{Name: "toolkit/enums", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "String-backed enums with labels"},
	{Name: "toolkit/hashid", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Obfuscated public IDs"},
	{Name: "toolkit/html", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "HTML helpers"},
	{Name: "toolkit/jsonschema", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "JSON Schema validation"},
	{Name: "toolkit/lock", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Process-local atomic locks"},
	{Name: "toolkit/markdown", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Markdown renderer"},
	{Name: "toolkit/money", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Money helpers"},
	{Name: "toolkit/num", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Number helpers"},
	{Name: "toolkit/process", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "OS process runner"},
	{Name: "toolkit/str", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "String helpers"},
	{Name: "toolkit/timing", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "Timing / stopwatch helpers"},
	{Name: "toolkit/zip", Layer: kernel.LayerAddon, Kind: kernel.KindLibrary, Description: "ZIP archive helpers"},
}
