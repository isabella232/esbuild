package snap_api

import (
	"testing"
)

var snapApiSuite = suite{
	name: "Snap API",
}

func TestEntryRequiringLocalModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			"/entry.js": `
				const { oneTwoThree } = require('./foo')
                module.exports = function () {
				  console.log(oneTwoThree)
			    }
			`,
			"/foo.js": `exports.oneTwoThree = 123`,
		},
		entryPoints: []string{"/entry.js"},
	},

		buildResult{
			files: map[string]string{
				`/entry.js`: `
let oneTwoThree;
function __get_oneTwoThree__() {
  return oneTwoThree = oneTwoThree || require_foo().oneTwoThree
}
module.exports = function() {
  get_console().log(__get_oneTwoThree__());
};`,
				`/foo.js`: `
var require_foo = __commonJS((exports2) => {
  exports2.oneTwoThree = 123;
});`,
			},
		},
	)
}

func TestEntryImportingLocalModule(t *testing.T) {
	snapApiSuite.expectBuild(t, built{
		files: map[string]string{
			"/entry.js": `
				import { oneTwoThree } from'./foo'
                module.exports = function () {
				  console.log(oneTwoThree)
			    }
			`,
			"/foo.js": `exports.oneTwoThree = 123`,
		},
		entryPoints: []string{"/entry.js"},
	},
		buildResult{
			files: map[string]string{
				`/foo.js`: `
var require_foo = __commonJS((exports2) => {
  exports2.oneTwoThree = 123;
});`,
				`/entry.js`: `
let foo;
function __get_foo__() {
  return foo = foo || __toModule(require_foo())
}
module.exports = function() {
  get_console().log(__get_foo__().oneTwoThree);
};`,
			},
		},
	)
}