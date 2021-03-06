package snap_printer

import "testing"

func TestIsolatedRequireRewrites(t *testing.T) {
	expectPrinted(t, "const foo = require('./foo')", `
let foo;
function __get_foo__() {
  return foo = foo || (require("./foo"))
}
`, ReplaceAll)

	expectPrinted(t, `
 const foo = require('./foo'),
   bar = require('./bar')
 `, `
let foo;
function __get_foo__() {
  return foo = foo || (require("./foo"))
}

let bar;
function __get_bar__() {
  return bar = bar || (require("./bar"))
}
`, ReplaceAll)
}

func TestIntegratedRequireRewrites(t *testing.T) {
	expectPrinted(t, `
const a = 1
const foo = require('./foo')
const b = 'hello world'
`, `
const a = 1;

let foo;
function __get_foo__() {
  return foo = foo || (require("./foo"))
}
const b = "hello world";
`, ReplaceAll)

	expectPrinted(t, `
const foo = require('./foo'),
  a = 1,
  bar = require('./bar'),
  b = 'hello world'
`, `
let foo;
function __get_foo__() {
  return foo = foo || (require("./foo"))
}
const a = 1;

let bar;
function __get_bar__() {
  return bar = bar || (require("./bar"))
}
const b = "hello world";`,
		ReplaceAll)
}

func TestRequireReferences(t *testing.T) {
	expectPrinted(t, `
const foo = require('./foo')
function logFoo() {
  console.log(foo.bar)
}
`, `
let foo;
function __get_foo__() {
  return foo = foo || (require("./foo"))
}
function logFoo() {
  get_console().log((__get_foo__()).bar);
}
`, ReplaceAll)
}

func TestSingleLateAssignment(t *testing.T) {
	expectPrinted(t, `
let a;
a = require('a')
`, `
let __get_a__;
let a;

__get_a__ = function() {
  return a = a || (require("a"))
};`, ReplaceAll)
}

func TestDoubleLateAssignment(t *testing.T) {
	expectPrinted(t, `
let a, b;
a = require('a')
b = require('b')
`, `
let __get_a__, __get_b__;
let a, b;

__get_a__ = function() {
  return a = a || (require("a"))
};

__get_b__ = function() {
  return b = b || (require("b"))
};
`, ReplaceAll)
}

func TestSingleLateAssignmentWithReference(t *testing.T) {
	expectPrinted(t, `
let a;
a = require('a')
`, `
let __get_a__;
let a;

__get_a__ = function() {
  return a = a || (require("a"))
};
`, ReplaceAll)
}

func TestDoubleLateAssignmentReplaceFilter(t *testing.T) {
	expectPrinted(t, `
let a, b;
a = require('a')
b = require('b')
`, `
let __get_a__;
let a, b;

__get_a__ = function() {
  return a = a || (require("a"))
};
b = require("b");
`, func(mod string) bool { return mod == "a" })
}

func TestConsoleReplacment(t *testing.T) {
	expectPrinted(
		t,
		`console.log('hello')`,
		`get_console().log("hello");`,
		ReplaceAll)
}

func TestProcessReplacement(t *testing.T) {
	expectPrinted(
		t,
		`process.a = 1`,
		`get_process().a = 1;`,
		ReplaceAll)
}

func TestReferencingGlobalProcessAndConstOfSameNamet(t *testing.T) {
	expectPrinted(
		t,
		`
{
  process.a = 1
}
{
  const process = {}
  process.b = 1
}
`, `
{
  get_process().a = 1;
}
{
  const process = {};
  process.b = 1;
}
`,
		ReplaceAll)
}

func TestRequireDeclPropertyChain(t *testing.T) {
	expectPrinted(t, `
const bar = require('foo').bar
`, `
let bar;
function __get_bar__() {
  return bar = bar || (require("foo").bar)
}
`, ReplaceAll)

	expectPrinted(t, `
const baz = require('foo').bar.baz
`, `
let baz;
function __get_baz__() {
  return baz = baz || (require("foo").bar.baz)
}
`, ReplaceAll)
}

func TestRequireLateAssignmentPropertyChain(t *testing.T) {
	expectPrinted(t, `
let bar
bar = require('foo').bar
`, `
let __get_bar__;
let bar;

__get_bar__ = function() {
  return bar = bar || (require("foo").bar)
};
`, ReplaceAll)

	expectPrinted(t, `
let baz
baz = require('foo').bar.baz
`, `
let __get_baz__;
let baz;

__get_baz__ = function() {
  return baz = baz || (require("foo").bar.baz)
};
`, ReplaceAll)
}

func TestDestructuringDeclarationReferenced(t *testing.T) {
	expectPrinted(t, `
const { foo, bar } = require('foo-bar')
function id() {
  foo.id = 'hello'
}
`, `
let foo;
function __get_foo__() {
  return foo = foo || (require("foo-bar").foo)
}

let bar;
function __get_bar__() {
  return bar = bar || (require("foo-bar").bar)
}
function id() {
  (__get_foo__()).id = "hello";
}
`, ReplaceAll)
}

func TestDestructuringLateAssignmentReferenced(t *testing.T) {
	expectPrinted(t, `
let foo, bar;
({ foo, bar } = require('foo-bar'))
function id() {
  foo.id = 'hello'
}
`, `
let __get_foo__, __get_bar__;
let foo, bar;

__get_foo__ = function() {
  return foo = foo || (require("foo-bar").foo)
}
__get_bar__ = function() {
  return bar = bar || (require("foo-bar").bar)
};
function id() {
  (__get_foo__()).id = "hello";
}
`, ReplaceAll)
}

func TestAssignToSameVarConditionallyAndReferenceIt(t *testing.T) {
	expectPrinted(t, `
let a
if (condition) {
  a = require('a')
} else { 
  a = require('b')
}
function foo() {
  a.b = 'c'
}
`, `
let __get_a__;
let a;
if (condition) {
  
__get_a__ = function() {
  return a = a || (require("a"))
};
} else {
  
__get_a__ = function() {
  return a = a || (require("b"))
};
}
function foo() {
  (__get_a__()).b = "c";
}
`, ReplaceAll)
}

func TestVarAssignedToRequiredVarAndReferenced(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const b = a
function main() {
  b.c = 1
}
`, `
let a;
function __get_a__() {
  return a = a || (require("a"))
}

let b;
function __get_b__() {
  return b = b || ((__get_a__()))
}
function main() {
  (__get_b__()).c = 1;
}
`, ReplaceAll)

}

func TestVarAssignedToPropertyOfRequiredVarAndReferenced(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const b = a.foo
function main() {
  b.c = 1
}
`, `
let a;
function __get_a__() {
  return a = a || (require("a"))
}

let b;
function __get_b__() {
  return b = b || ((__get_a__()).foo)
}
function main() {
  (__get_b__()).c = 1;
}
`, ReplaceAll)

}
func TestDestructuredVarsAssignedToPropertyOfRequiredVarAndReferenced(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
const { foo, bar } = a
function main() {
  return foo + bar 
}
`, `
let a;
function __get_a__() {
  return a = a || (require("a"))
}

let foo;
function __get_foo__() {
  return foo = foo || ((__get_a__()).foo)
}

let bar;
function __get_bar__() {
  return bar = bar || ((__get_a__()).bar)
}
function main() {
  return (__get_foo__()) + (__get_bar__());
}
`, ReplaceAll)
}

func TestVarsInSingleDeclarationReferencingEachOtherReferenced(t *testing.T) {
	expectPrinted(t, `
let a = require('a'), b  = a.c
function main() {
  return a + b 
}
`, `
let a;
function __get_a__() {
  return a = a || (require("a"))
}

let b;
function __get_b__() {
  return b = b || ((__get_a__()).c)
}
function main() {
  return (__get_a__()) + (__get_b__());
}
`, ReplaceAll)
}

func TestLateAssignmentToRequireReference(t *testing.T) {
	expectPrinted(t, `
const a = require('a')
let b
b = a.c
function main() {
  return a + b 
}
`, `
let __get_b__;

let a;
function __get_a__() {
  return a = a || (require("a"))
}
let b;

__get_b__ = function() {
  return b = b || ((__get_a__()).c)
};
function main() {
  return (__get_a__()) + (__get_b__());
}
`, ReplaceAll)
}

func TestIndirectReferencesToRequireInSameDeclaration(t *testing.T) {
	expectPrinted(t, `
let d = require("d"), e = d.e, f = e.f;
`, `
let d;
function __get_d__() {
  return d = d || (require("d"))
}

let e;
function __get_e__() {
  return e = e || ((__get_d__()).e)
}

let f;
function __get_f__() {
  return f = f || ((__get_e__()).f)
}
`, ReplaceAll)
}

func TestIndirectReferencesToRequireLateAssign(t *testing.T) {
	expectPrinted(t, `
let d, e, f;
d = require("d");
e = d.e;
f = e.f;
`, `
let __get_d__, __get_e__, __get_f__;
let d, e, f;

__get_d__ = function() {
  return d = d || (require("d"))
};

__get_e__ = function() {
  return e = e || ((__get_d__()).e)
};

__get_f__ = function() {
  return f = f || ((__get_e__()).f)
}; `, ReplaceAll)
}

func TestDeclarationToCallResultWithRequireReferenceArgReferenced(t *testing.T) {
	expectPrinted(t, `
var pack = require('pack')
const x = someCall(pack);
function main() {
  return x + 1
}
`, `
let pack;
function __get_pack__() {
  return pack = pack || (require("pack"))
}

let x;
function __get_x__() {
  return x = x || (someCall((__get_pack__())))
}
function main() {
  return (__get_x__()) + 1;
}
`, ReplaceAll)
}

func TestDeclarationWithEBinaryReferencingRequire(t *testing.T) {
	expectPrinted(t, `
const c = require('c').foo.bar
const d = c.X | c.Y | c.Z
`, `
let c;
function __get_c__() {
  return c = c || (require("c").foo.bar)
}

let d;
function __get_d__() {
  return d = d || ((__get_c__()).X | (__get_c__()).Y | (__get_c__()).Z)
}
`, ReplaceAll)
}

func TestTopLevelVsNestedRequiresAndReferences(t *testing.T) {
	expectPrinted(t, `
function nested() {
  const a = require('a')
}
const b = require('b')
const c = b.foo
`, `
function nested() {
  const a = require("a");
}

let b;
function __get_b__() {
  return b = b || (require("b"))
}

let c;
function __get_c__() {
  return c = c || ((__get_b__()).foo)
}
`, ReplaceAll)
}

func TestLateAssignedTopLevelVsNestedRequiresAndReferences(t *testing.T) {
	expectPrinted(t, `
function nested() {
  let a
  a = require('a')
}
let b, c
b = require('b')
c = b.foo
`, `
let __get_b__, __get_c__;
function nested() {
  let a;
  a = require("a");
}
let b, c;

__get_b__ = function() {
  return b = b || (require("b"))
};

__get_c__ = function() {
  return c = c || ((__get_b__()).foo)
};
`, ReplaceAll)
}

func TestRequireReferencesInsideBlock(t *testing.T) {
	expectPrinted(t, `
{
  const a = require('a')
  const c = a.bar
}
`, `
{

let a;
function __get_a__() {
  return a = a || (require("a"))
}

let c;
function __get_c__() {
  return c = c || ((__get_a__()).bar)
}
}
`, ReplaceAll)
}

func TestRequireWithCallchain(t *testing.T) {
	expectPrinted(t, `
 var debug = require('debug')('express:view')
`, `
let debug;
function __get_debug__() {
  return debug = debug || (require("debug")("express:view"))
}
`, ReplaceAll)

	expectPrinted(t, `
 var chain = require('chainer')('hello')('world')(foo())(1)
`, `
let chain;
function __get_chain__() {
  return chain = chain || (require("chainer")("hello")("world")(foo())(1))
}
`, ReplaceAll)
}

func TestRequireWithCallchainAndPropChain(t *testing.T) {
	expectPrinted(t, `
 var chain = require('chainer')('hello').foo.bar
`, `
let chain;
function __get_chain__() {
  return chain = chain || (require("chainer")("hello").foo.bar)
}
`, ReplaceAll)
}

func TestDeclarationReferencingGlobal(t *testing.T) {
	expectPrinted(t, `
const { relative } = require('path')
var basePath = process.cwd()
function relToBase(s) {
  return relative(basePath, relative)
}
`, `
let relative;
function __get_relative__() {
  return relative = relative || (require("path").relative)
}

let basePath;
function __get_basePath__() {
  return basePath = basePath || (get_process().cwd())
}
function relToBase(s) {
  return (__get_relative__())((__get_basePath__()), (__get_relative__()));
}
`, ReplaceAll)
}

func TestPropChainWithCalledProperty(t *testing.T) {
	expectPrinted(t, `
var tmpDir = require('os').tmpdir();
`, `
let tmpDir;
function __get_tmpDir__() {
  return tmpDir = tmpDir || (require("os").tmpdir())
}
`, ReplaceAll)
}

func TestMultiDeclarationsLastReplaced(t *testing.T) {
	expectPrinted(t,
		`
  var p = null,
    q = null,
    u = Date.now()
`, `
var p = null;
var q = null;

let u;
function __get_u__() {
  return u = u || (Date.now())
}
`, ReplaceAll)
}

// TODO: do we care that the below rewrite doesn't work as expected
// since types is defined initially, i.e. we'll never run Object.create(null)?
// Latter excludes `Object` prototype methods.
// Example from `esprima/esprima.js`
func TestDeclareReplacementsMultipleUseStrict(t *testing.T) {
	expectPrinted(t,
		`
(function (root, factory) {
  'use strict';
  var typesOuter = {}
  if (typeof Object.create === 'function') {
  	typesOuter = Object.create(null);
  }
}(this, function (exports) {
    'use strict';

	var types = {};

	if (typeof Object.create === 'function') {
		types = Object.create(null);
	}
}))
`, `
(function(root, factory) {
  "use strict";
let __get_typesOuter__;
  var typesOuter = {};
  if (typeof Object.create === "function") {
    
__get_typesOuter__ = function() {
  return typesOuter = typesOuter || (Object.create(null))
};
  }
})(this, function(exports) {
  "use strict";
let __get_types__;
  var types = {};
  if (typeof Object.create === "function") {
    
__get_types__ = function() {
  return types = types || (Object.create(null))
};
  }
});
`, ReplaceAll)

}

func TestReplacePartOfLogicalAnd(t *testing.T) {
	// Replacement here is too complex to do correctly.
	// In the below example __get_ke__ would exist conditionally and cause problems when
	// invoked if not.
	// Until we encounter a case where this is necessary we just leave it unchanged.
	// @see handleEBinary
	expectPrinted(t,
		`
var ya
var ke = null
ya && 'documentMode' in document && (ke = document.documentMode)
`, `

var ya;
var ke = null;
ya && "documentMode" in get_document() && (ke = get_document().documentMode);
`, ReplaceAll)
}

func TestWrappingConstructorReferencedViaNew(t *testing.T) {
	expectPrinted(t,
		`
var EE = require('events')
const emitter= new EE()
`, `
let EE;
function __get_EE__() {
  return EE = EE || (require("events"))
}
const emitter = new (__get_EE__())();
`, ReplaceAll)

}

func TestWrappingConstructorReferencedViaNewAssigned(t *testing.T) {
	expectPrinted(t,
		`
var EE = require('events')
var emitter
emitter = process.__signal_exit_emitter__ = new EE()
`, `
let __get_emitter__;

let EE;
function __get_EE__() {
  return EE = EE || (require("events"))
}
var emitter;

__get_emitter__ = function() {
  return emitter = emitter || (get_process().__signal_exit_emitter__ = new (__get_EE__())())
};
`, ReplaceAll)

}

func TestGlobalRewriteAsPartOfDeclChain(t *testing.T) {
	expectPrinted(t,
		`
	var o = Object, keysShim;
	keysShim = 1;
`, `
let o;
function __get_o__() {
  return o = o || (Object)
}
var keysShim;
keysShim = 1;
`, ReplaceAll)

}

func TestGetterForExportsAssignmentOfDeferred(t *testing.T) {
	expectPrinted(t,
		`
const res = require('./lib/response')
exports.response = res
`, `
let res;
function __get_res__() {
  return res = res || (require("./lib/response"))
}
Object.defineProperty(exports, "response", { get: () => (__get_res__()) });
`, ReplaceAll)
}

func TestGetterForModuleExportsAssignmentOfDeferred(t *testing.T) {
	expectPrinted(t,
		`
const res = require('./lib/response')
module.exports.response = res
`, `
let res;
function __get_res__() {
  return res = res || (require("./lib/response"))
}
Object.defineProperty(module.exports, "response", { get: () => (__get_res__()) });
`, ReplaceAll)
}

func TestGetterForExportsOfDirectRequire(t *testing.T) {
	expectPrinted(t,
		`
exports.fs = require('fs')
`, `
Object.defineProperty(exports, "fs", { get: () => require("fs") });
`, ReplaceAll)
}

func TestGetterForIndexedExports(t *testing.T) {
	expectPrinted(t,
		`
const res = require('./lib/response')
exports.response = res
exports['fs'] = require('fs')

`, `
let res;
function __get_res__() {
  return res = res || (require("./lib/response"))
}
Object.defineProperty(exports, "response", { get: () => (__get_res__()) });
Object.defineProperty(exports, "fs", { get: () => require("fs") });
`, ReplaceAll)
}

func TestNotWrappingInsideArrowFunction(t *testing.T) {
	// First case doesn't wrap since it is already
	expectPrinted(t,
		`
const x = () => {
  const options = process.cwd()
  return options
}
`, `
const x = () => {
  const options = get_process().cwd();
  return options;
};
`, ReplaceAll)

	// Second case does wrap as usual
	expectPrinted(t,
		`
const options = process.cwd()
module.exports = () => options
`, `
let options;
function __get_options__() {
  return options = options || (get_process().cwd())
}
module.exports = () => (__get_options__());
`, ReplaceAll)

}

func TestAssigningToPreviouslyDeferredDoesNotWrapAssignee(t *testing.T) {
	expectPrinted(t,
		`
function fallback() {}
var _defer = process
_defer = fallback
`, `
function fallback() {
}

let _defer;
function __get__defer__() {
  return _defer = _defer || (get_process())
}
_defer = fallback;
`, ReplaceAll)

}

func TestRequireWithPropertyCallIncludingArgs(t *testing.T) {
	expectPrinted(t,
		`
const p = require('editions').requirePackage(__dirname, require)
`, `
let p;
function __get_p__() {
  return p = p || (require("editions").requirePackage(__dirname require))
}
`, ReplaceAll)

}

func TestDuplicateRequireDeclaration(t *testing.T) {
	expectPrinted(t,
		`
var Buffer = require('buffer').Buffer
var Buffer = require('buffer').Buffer
`,
		`
let Buffer;
function __get_Buffer__() {
  return Buffer = Buffer || (require("buffer").Buffer)
}
`, ReplaceAll)
}

func TestInvokedRequireAlwaysDeferred(t *testing.T) {
	expectPrinted(t, `
 var d1 = require('invoked')('hello')
 var d2 = require('not-invoked')
`, `
let d1;
function __get_d1__() {
  return d1 = d1 || (require("invoked")("hello"))
}
var d2 = require("not-invoked");
`, ReplaceNone)
}

/*
# Incorrectly handled cases.

Below are cases that are currently not handled 100% correctly.
They don't all need to be addressed as in some instances they don't cause problems.

## Conditionally reassignments

// Below won't cause issues  since in latest Node.js `require('events')` returns a
// function.
```
var EE = require('events')
if (typeof EE !== 'function') {
  EE = EE.EventEmitter
}
```
becomes:
```
let EE;
function __get_EE__() {
  return EE = EE || require("events")
}
if (typeof __get_EE__() !== "function") {
  // Overwriting the original __get_EE__ here which is invoked below resulting
  // in recursion.
  __get_EE__ = function() {
     return EE = EE || __get_EE__().EventEmitter
  };
}
```

*/
