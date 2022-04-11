package js

import (
	"bytes"
	"sort"

	"github.com/tdewolff/parse/v2/js"
)

type renamer struct {
	identStart    []byte
	identContinue []byte
	identOrder    map[byte]int
	reserved      map[string]struct{}
	rename        bool
}

func newRenamer(rename, useCharFreq bool) *renamer {
	reserved := make(map[string]struct{}, len(js.Keywords))
	for name := range js.Keywords {
		reserved[name] = struct{}{}
	}
	identStart := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$")
	identContinue := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$0123456789")
	if useCharFreq {
		// sorted based on character frequency of a collection of JS samples (incl. the var names!)
		identStart = []byte("etnsoiarclduhmfpgvbjy_wOxCEkASMFTzDNLRPHIBV$WUKqYGXQZJ")
		identContinue = []byte("etnsoiarcldu14023hm8f6pg57v9bjy_wOxCEkASMFTzDNLRPHIBV$WUKqYGXQZJ")
	}
	identOrder := map[byte]int{}
	for i, c := range identStart {
		identOrder[c] = i
	}
	return &renamer{
		identStart:    identStart,
		identContinue: identContinue,
		identOrder:    identOrder,
		reserved:      reserved,
		rename:        rename,
	}
}

func (r *renamer) renameScope(scope js.Scope) {
	if !r.rename {
		return
	}

	i := 0
	sort.Sort(js.VarsByUses(scope.Declared))
	for _, v := range scope.Declared {
		v.Data = r.getName(v.Data, i)
		i++
		for r.isReserved(v.Data, scope.Undeclared) {
			v.Data = r.getName(v.Data, i)
			i++
		}
	}
}

func (r *renamer) isReserved(name []byte, undeclared js.VarArray) bool {
	if 1 < len(name) { // there are no keywords or known globals that are one character long
		if _, ok := r.reserved[string(name)]; ok {
			return true
		}
	}
	for _, v := range undeclared {
		for v.Link != nil {
			v = v.Link
		}
		if bytes.Equal(v.Data, name) {
			return true
		}
	}
	return false
}

func (r *renamer) getIndex(name []byte) int {
	index := 0
NameLoop:
	for i, b := range name {
		chars := r.identContinue
		if i == 0 {
			chars = r.identStart
		} else {
			index *= len(r.identContinue)
		}
		for j, c := range chars {
			if b == c {
				index += j
				continue NameLoop
			}
		}
		return -1
	}
	for n := 0; n < len(name)-1; n++ {
		offset := len(r.identStart)
		for i := 0; i < n; i++ {
			offset *= len(r.identContinue)
		}
		index += offset
	}
	return index
}

func (r *renamer) getName(name []byte, index int) []byte {
	// Generate new names for variables where the last character is (a-zA-Z$_) and others are (a-zA-Z).
	// Thus we can have 54 one-character names and 52*54=2808 two-character names for every branch leaf.
	// That is sufficient for virtually all input.

	if index < len(r.identStart) {
		name[0] = r.identStart[index]
		return name[:1]
	}

	index -= len(r.identStart)
	n := 2
	for {
		offset := len(r.identStart)
		for i := 0; i < n-1; i++ {
			offset *= len(r.identContinue)
		}
		if index < offset {
			break
		}
		index -= offset
		n++
	}

	if cap(name) < n {
		name = make([]byte, n)
	} else {
		name = name[:n]
	}
	for j := n - 1; 0 < j; j-- {
		name[j] = r.identContinue[index%len(r.identContinue)]
		index /= len(r.identContinue)
	}
	name[0] = r.identStart[index]
	return name
}

////////////////////////////////////////////////////////////////

func hasDefines(v *js.VarDecl) bool {
	for _, item := range v.List {
		if item.Default != nil {
			return true
		}
	}
	return false
}

func bindingVars(ibinding js.IBinding) (vs []*js.Var) {
	switch binding := ibinding.(type) {
	case *js.Var:
		vs = append(vs, binding)
	case *js.BindingArray:
		for _, item := range binding.List {
			if item.Binding != nil {
				vs = append(vs, bindingVars(item.Binding)...)
			}
		}
		if binding.Rest != nil {
			vs = append(vs, bindingVars(binding.Rest)...)
		}
	case *js.BindingObject:
		for _, item := range binding.List {
			if item.Value.Binding != nil {
				vs = append(vs, bindingVars(item.Value.Binding)...)
			}
		}
		if binding.Rest != nil {
			vs = append(vs, binding.Rest)
		}
	}
	return
}

func addDefinition(decl *js.VarDecl, binding js.IBinding, value js.IExpr, forward bool) {
	// see if not already defined in variable declaration list
	// if forward is set, binding=value comes before decl, otherwise the reverse holds true
	vars := bindingVars(binding)

	// remove variables in destination
RemoveVarsLoop:
	for _, vbind := range vars {
		for i, item := range decl.List {
			if v, ok := item.Binding.(*js.Var); ok && item.Default == nil && v == vbind {
				v.Uses--
				decl.List = append(decl.List[:i], decl.List[i+1:]...)
				continue RemoveVarsLoop
			}
		}

		if value != nil {
			// variable declaration must be somewhere else, find and remove it
			for _, decl2 := range decl.Scope.Func.VarDecls {
				for i, item := range decl2.List {
					if v, ok := item.Binding.(*js.Var); ok && item.Default == nil && v == vbind {
						v.Uses--
						decl2.List = append(decl2.List[:i], decl2.List[i+1:]...)
						continue RemoveVarsLoop
					}
				}
			}
		}
	}

	// add declaration to destination
	item := js.BindingElement{Binding: binding, Default: value}
	if forward {
		decl.List = append([]js.BindingElement{item}, decl.List...)
	} else {
		decl.List = append(decl.List, item)
	}
}

func mergeVarDecls(dst, src *js.VarDecl, forward bool) {
	// this is the second VarDecl, so we are hoisting var declarations, which means the forInit variables are already in 'left'
	for j := 0; j < len(src.List); j++ {
		addDefinition(dst, src.List[j].Binding, src.List[j].Default, forward)
		src.List = append(src.List[:j], src.List[j+1:]...)
		j--
	}
	// remove from vardecls list of scope
	scope := src.Scope.Func
	for i, decl := range scope.VarDecls {
		if src == decl {
			scope.VarDecls = append(scope.VarDecls[:i], scope.VarDecls[i+1:]...)
			break
		}
	}
}

func mergeVarDeclExprStmt(decl *js.VarDecl, exprStmt *js.ExprStmt, forward bool) bool {
	if src, ok := exprStmt.Value.(*js.VarDecl); ok {
		// this happens when a variable declarations is converted to an expression
		mergeVarDecls(decl, src, forward)
		return true
	} else if commaExpr, ok := exprStmt.Value.(*js.CommaExpr); ok {
		n := 0
		for i := 0; i < len(commaExpr.List); i++ {
			item := commaExpr.List[i]
			if forward {
				item = commaExpr.List[len(commaExpr.List)-i-1]
			}
			if src, ok := item.(*js.VarDecl); ok {
				// this happens when a variable declarations is converted to an expression
				mergeVarDecls(decl, src, forward)
				n++
				continue
			} else if binaryExpr, ok := item.(*js.BinaryExpr); ok && binaryExpr.Op == js.EqToken {
				if v, ok := binaryExpr.X.(*js.Var); ok && v.Decl == js.VariableDecl {
					addDefinition(decl, v, binaryExpr.Y, forward)
					n++
					continue
				}
			}
			break
		}
		merge := n == len(commaExpr.List)
		if !forward {
			commaExpr.List = commaExpr.List[n:]
		} else {
			commaExpr.List = commaExpr.List[:len(commaExpr.List)-n]
		}
		return merge
	} else if binaryExpr, ok := exprStmt.Value.(*js.BinaryExpr); ok && binaryExpr.Op == js.EqToken {
		if v, ok := binaryExpr.X.(*js.Var); ok && v.Decl == js.VariableDecl {
			addDefinition(decl, v, binaryExpr.Y, forward)
			return true
		}
	}
	return false
}

func (m *jsMinifier) countHoistLength(ibinding js.IBinding) int {
	if !m.o.KeepVarNames {
		return len(bindingVars(ibinding)) * 2 // assume that var name will be of length one, +1 for the comma
	}

	n := 0
	for _, v := range bindingVars(ibinding) {
		n += len(v.Data) + 1 // +1 for the comma when added to other declaration
	}
	return n
}

func (m *jsMinifier) hoistVars(body *js.BlockStmt) {
	// Hoist all variable declarations in the current module/function scope to the top.
	// If the first statement is a var declaration, expand it. Otherwise prepend a new var declaration.
	// Except for the first var declaration, all others are converted to expressions. This is possible because an ArrayBindingPattern and ObjectBindingPattern can be converted to an ArrayLiteral or ObjectLiteral respectively, as they are supersets of the BindingPatterns.
	if 1 < len(body.Scope.VarDecls) {
		// Select which variable declarations will be hoisted (convert to expression) and which not
		best := 0
		score := make([]int, len(body.Scope.VarDecls)) // savings if hoisted
		hoist := make([]bool, len(body.Scope.VarDecls))
		for i, varDecl := range body.Scope.VarDecls {
			hoist[i] = true
			score[i] = 4 // "var "
			if !varDecl.InForInOf {
				n := 0
				nArrays := 0
				nObjects := 0
				hasDefinitions := false
				for j, item := range varDecl.List {
					if item.Default != nil {
						if _, ok := item.Binding.(*js.BindingObject); ok {
							if j != 0 && nArrays == 0 && nObjects == 0 {
								varDecl.List[0], varDecl.List[j] = varDecl.List[j], varDecl.List[0]
							}
							nObjects++
						} else if _, ok := item.Binding.(*js.BindingArray); ok {
							if j != 0 && nArrays == 0 && nObjects == 0 {
								varDecl.List[0], varDecl.List[j] = varDecl.List[j], varDecl.List[0]
							}
							nArrays++
						}
						score[i] -= m.countHoistLength(item.Binding) // var names and commas
						hasDefinitions = true
						n++
					}
				}
				if !hasDefinitions {
					score[i] = 5 - 1 // 1 for a comma
					if varDecl.InFor {
						score[i]-- // semicolon can be reused
					}
				}
				if nObjects != 0 && !varDecl.InFor && nObjects == n {
					score[i] -= 2 // required parenthesis around braces
				}
				if nArrays != 0 || nObjects != 0 {
					score[i]-- // space after var disappears
				}
				if score[i] < score[best] || body.Scope.VarDecls[best].InForInOf {
					// select var decl with the least savings if hoisted
					best = i
				}
				if score[i] < 0 {
					hoist[i] = false
				}
			}
		}
		if body.Scope.VarDecls[best].InForInOf {
			// no savings possible
			return
		}

		decl := body.Scope.VarDecls[best]
		hoist[best] = false

		// get original declarations
		orig := []*js.Var{}
		for _, item := range decl.List {
			orig = append(orig, bindingVars(item.Binding)...)
		}

		// hoist other variable declarations in this function scope but don't initialize yet
		j := 0
		for i, varDecl := range body.Scope.VarDecls {
			if hoist[i] {
				varDecl.TokenType = js.ErrorToken
				for _, item := range varDecl.List {
					refs := bindingVars(item.Binding)
					bindingElements := make([]js.BindingElement, 0, len(refs))
				DeclaredLoop:
					for _, ref := range refs {
						for _, v := range orig {
							if ref == v {
								continue DeclaredLoop
							}
						}
						bindingElements = append(bindingElements, js.BindingElement{Binding: ref, Default: nil})
						orig = append(orig, ref)

						s := decl.Scope
						for s != nil && s != s.Func {
							s.AddUndeclared(ref)
							s = s.Parent
						}
						if item.Default != nil {
							ref.Uses++
						}
					}
					if i < best {
						// prepend
						decl.List = append(decl.List[:j], append(bindingElements, decl.List[j:]...)...)
						j += len(bindingElements)
					} else {
						// append
						decl.List = append(decl.List, bindingElements...)
					}
				}
			}
		}

		// rearrange to put array/object first
		var prevRefs []*js.Var
	BeginArrayObject:
		for i, item := range decl.List {
			refs := bindingVars(item.Binding)
			if _, ok := item.Binding.(*js.Var); !ok {
				if i != 0 {
					interferes := false
					if item.Default != nil {
					InterferenceLoop:
						for _, ref := range refs {
							for _, v := range prevRefs {
								if ref == v {
									interferes = true
									break InterferenceLoop
								}
							}
						}
					}
					if !interferes {
						decl.List[0], decl.List[i] = decl.List[i], decl.List[0]
						break BeginArrayObject
					}
				} else {
					break BeginArrayObject
				}
			}
			if item.Default != nil {
				prevRefs = append(prevRefs, refs...)
			}
		}
	}
}
