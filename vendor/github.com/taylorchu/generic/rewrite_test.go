package generic

import "testing"

func TestRewritePackage(t *testing.T) {
	testRewritePackage(t, "github.com/taylorchu/generic/fixture/basic", "result", map[string]Target{
		"Type": Target{Ident: "int64"},
	}, "output/basic")
}

func TestRewritePackageVendoring(t *testing.T) {
	testRewritePackage(t, "github.com/taylorchu/generic/fixture/basic", "result", map[string]Target{
		"Type": Target{Ident: "vendoring.Number", Import: "github.com/taylorchu/generic/fixture/vendoring"},
	}, "output/vendoring")
}

func TestRewritePackageMethod(t *testing.T) {
	testRewritePackage(t, "github.com/taylorchu/generic/fixture/method", "result", map[string]Target{
		"Type2": Target{Ident: "generic.Target", Import: "github.com/taylorchu/generic"},
	}, "output/method")
}

func TestRewritePackageInternal(t *testing.T) {
	testRewritePackage(t, "github.com/taylorchu/generic/fixture/basic", "internal/result", map[string]Target{
		"Type": Target{Ident: "int64"},
	}, "output/internal")
}

func TestRewritePackageDotRename(t *testing.T) {
	testRewritePackage(t, "github.com/taylorchu/generic/fixture/rename", ".result", map[string]Target{
		"Type": Target{Ident: "int64"},
	}, "output/dot_rename")
}

func TestRewritePackageQueue(t *testing.T) {
	testRewritePackage(t, "github.com/taylorchu/generic/fixture/queue", "result", map[string]Target{
		"Type":      Target{Ident: "int64"},
		"TypeQueue": Target{Ident: "FIFO"},
	}, "output/queue")
}

func TestRewritePackageDotQueue(t *testing.T) {
	testRewritePackageWithInput(t, "github.com/taylorchu/generic/fixture/queue", ".result", map[string]Target{
		"Type":      Target{Ident: "Data"},
		"TypeQueue": Target{Ident: "FIFO"},
	},
		"input/data",
		"output/dot_queue",
	)
}

func TestRewritePackageDotQueuePrefix(t *testing.T) {
	testRewritePackageWithInput(t, "github.com/taylorchu/generic/fixture/queue", ".result", map[string]Target{
		"Type": Target{Ident: "Data"},
	},
		"input/data",
		"output/dot_queue_prefix",
	)
}

func TestRewritePackageDotContainer(t *testing.T) {
	testRewritePackageWithInput(t, "github.com/taylorchu/generic/fixture/container", ".result", map[string]Target{
		"Type":          Target{Ident: "*Data"},
		"TypeContainer": Target{Ident: "Box"},
	},
		"input/data",
		"output/dot_container",
	)
}

func TestRewritePackageDotContainerUpdate(t *testing.T) {
	testRewritePackageWithInput(t, "github.com/taylorchu/generic/fixture/container", ".result", map[string]Target{
		"Type":          Target{Ident: "*Data"},
		"TypeContainer": Target{Ident: "Box"},
	},
		"input/container_updated",
		"output/dot_container",
	)
}

func TestRewritePackageDotRenameUnresolved(t *testing.T) {
	testRewritePackageWithInput(t, "github.com/taylorchu/generic/fixture/rename", ".result", map[string]Target{
		"Type": Target{Ident: "Data"},
	},
		"input/data_unresolved",
		"output/dot_rename_unresolved",
	)
}
