package wizcompiler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/fasterthanlime/wizardry/wizardry/wizparser"
	"github.com/go-errors/errors"
)

type indentCallback func()

// Compile generates go code from a spellbook
func Compile(book wizparser.Spellbook) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, 0)
	}

	fullPath := filepath.Join(wd, "wizardry", "wizbook", "book.go")

	f, err := os.Create(fullPath)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	fmt.Println("Generating into:", fullPath)

	defer f.Close()

	lf := []byte("\n")
	oneIndent := []byte("  ")
	indentLevel := 0

	indent := func() {
		indentLevel++
	}

	outdent := func() {
		indentLevel--
	}

	withIndent := func(f indentCallback) {
		indent()
		f()
		outdent()
	}

	emit := func(format string, args ...interface{}) {
		if format != "" {
			for i := 0; i < indentLevel; i++ {
				f.Write(oneIndent)
			}
			fmt.Fprintf(f, format, args...)
		}
		f.Write(lf)
	}

	emit("// this file has been generated by github.com/fasterthanlime/wizardry")
	emit("// from a set of magic rules. you probably don't want to edit it by hand")
	emit("")

	emit("package wizbook")
	emit("")
	emit("import (")
	withIndent(func() {
		emit(strconv.Quote("fmt"))
		emit(strconv.Quote("encoding/binary"))
		emit(strconv.Quote("github.com/fasterthanlime/wizardry/wizardry"))
	})
	emit(")")
	emit("")

	emit("// silence import errors, if we don't use string/search etc.")
	emit("var _ wizardry.StringTestFlags")
	emit("var _ fmt.State")

	emit("var le binary.ByteOrder = binary.LittleEndian")
	emit("var be binary.ByteOrder = binary.BigEndian")
	for _, byteWidth := range []byte{1, 2, 4, 8} {
		emit("type i%d int%d", byteWidth*8, byteWidth*8)
		emit("type u%d uint%d", byteWidth*8, byteWidth*8)
	}
	emit("")

	for _, byteWidth := range []byte{1, 2, 4, 8} {
		for _, endianness := range []wizparser.Endianness{wizparser.LittleEndian, wizparser.BigEndian} {
			retType := fmt.Sprintf("u%d", byteWidth*8)

			emit("func readU%d%s(tb []byte, off i64) (%s, bool) {", byteWidth*8, endiannessString(endianness, false), retType)
			withIndent(func() {
				emit("if i64(len(tb)) < off+%d {", byteWidth)
				withIndent(func() {
					emit("return 0, false")
				})
				emit("}")

				if byteWidth == 1 {
					emit("pi := %s(tb[off])", retType)
				} else {
					emit("pi := %s.Uint%d(tb[off:])", endiannessString(endianness, false), byteWidth*8)
				}
				emit("return %s(pi), true", retType)
			})
			emit("}")
			emit("")
		}
	}

	currentLevel := 0

	var pages []string
	for page := range book {
		pages = append(pages, page)
	}
	sort.Strings(pages)

	for _, swapEndian := range []bool{false, true} {
		for _, page := range pages {
			rules := book[page]

			emit("func Identify%s(tb []byte, pageOff i64) ([]string, error) {", pageSymbol(page, swapEndian))
			withIndent(func() {
				emit("var out []string")
				emit("var gof i64") // globalOffset
				emit("gof &= gof")
				emit("var off i64") // lookupOffset
				emit("var ml i64")  // matchLength

				maxLevel := 0
				for _, rule := range rules {
					if rule.Level > maxLevel && len(rule.Description) > 0 {
						maxLevel = rule.Level
					}
				}
				for i := 0; i <= maxLevel; i++ {
					emit("m%d := false", i)
					emit("m%d = !!m%d", i, i)
				}
				emit("")

				ruleIndex := 0

				for {
					if ruleIndex >= len(rules) {
						break
					}
					rule := rules[ruleIndex]
					ruleIndex++

					for currentLevel < rule.Level {
						emit("if m%d {", currentLevel)
						currentLevel++
						indent()
					}

					for currentLevel > rule.Level {
						currentLevel--
						outdent()
						emit("}")

						emit("m%d = false", currentLevel)
					}

					emit("// %s", rule.Line)

					needsBreak := false
					if rule.Offset.OffsetType == wizparser.OffsetTypeIndirect {
						needsBreak = true
					}

					emitted := true

					if needsBreak {
						emit("rule%d:", ruleIndex)
						indent()
						emit("for {")
						indent()
					}

					switch rule.Offset.OffsetType {
					case wizparser.OffsetTypeDirect:
						if rule.Offset.IsRelative {
							emit("off = pageOff + gof + %s", quoteNumber(rule.Offset.Direct))
						} else {
							emit("off = pageOff + %s", quoteNumber(rule.Offset.Direct))
						}
					case wizparser.OffsetTypeIndirect:
						indirect := rule.Offset.Indirect

						emit("{")
						withIndent(func() {
							offsetAddress := quoteNumber(indirect.OffsetAddress)
							if indirect.IsRelative {
								offsetAddress = fmt.Sprintf("(gof + %s)", offsetAddress)
							}

							emit("ra, ok := readU%d%s(tb, %s)",
								indirect.ByteWidth*8,
								endiannessString(indirect.Endianness, swapEndian),
								offsetAddress)
							emit("if !ok { break rule%d }", ruleIndex)
							offsetAdjustValue := quoteNumber(indirect.OffsetAdjustmentValue)

							if indirect.OffsetAdjustmentIsRelative {
								offsetAdjustAddress := fmt.Sprintf("%s + %s", offsetAddress, quoteNumber(indirect.OffsetAdjustmentValue))
								emit("rb, ok := readU%d%s(tb, %s)",
									indirect.ByteWidth*8,
									endiannessString(indirect.Endianness, swapEndian),
									offsetAdjustAddress)
								emit("if !ok { break rule%d }", ruleIndex)
								offsetAdjustValue = "i64(rb)"
							}

							emit("off = i64(ra)")

							switch indirect.OffsetAdjustmentType {
							case wizparser.AdjustmentAdd:
								emit("off = off + %s", offsetAdjustValue)
							case wizparser.AdjustmentDiv:
								emit("off = off / %s", offsetAdjustValue)
							case wizparser.AdjustmentMul:
								emit("off = off * %s", offsetAdjustValue)
							case wizparser.AdjustmentSub:
								emit("off = off * %s", quoteNumber(indirect.OffsetAdjustmentValue))
							}

							if rule.Offset.IsRelative {
								emit("off += gof")
							}
						})
						emit("}")
					}

					switch rule.Kind.Family {
					case wizparser.KindFamilyInteger:
						ik, _ := rule.Kind.Data.(*wizparser.IntegerKind)

						if ik.MatchAny {
							emit("ml = %d", ik.ByteWidth)
						} else {
							emit("{")
							withIndent(func() {
								emit("iv, ok := readU%d%s(tb, %s)",
									ik.ByteWidth*8,
									endiannessString(ik.Endianness, swapEndian),
									"off",
								)

								lhs := "iv"

								operator := "=="
								switch ik.IntegerTest {
								case wizparser.IntegerTestEqual:
									operator = "=="
								case wizparser.IntegerTestNotEqual:
									operator = "!="
								case wizparser.IntegerTestLessThan:
									operator = "<"
								case wizparser.IntegerTestGreaterThan:
									operator = ">"
								}

								if ik.IntegerTest == wizparser.IntegerTestGreaterThan || ik.IntegerTest == wizparser.IntegerTestLessThan {
									lhs = fmt.Sprintf("i64(i%d(iv))", ik.ByteWidth*8)
								} else {
									lhs = fmt.Sprintf("u64(iv)")
								}

								if ik.DoAnd {
									lhs = fmt.Sprintf("%s&%s", lhs, quoteNumber(int64(ik.AndValue)))
								}

								switch ik.AdjustmentType {
								case wizparser.AdjustmentAdd:
									lhs = fmt.Sprintf("(%s+%s)", lhs, quoteNumber(ik.AdjustmentValue))
								case wizparser.AdjustmentSub:
									lhs = fmt.Sprintf("(%s-%s)", lhs, quoteNumber(ik.AdjustmentValue))
								case wizparser.AdjustmentMul:
									lhs = fmt.Sprintf("(%s*%s)", lhs, quoteNumber(ik.AdjustmentValue))
								case wizparser.AdjustmentDiv:
									lhs = fmt.Sprintf("(%s/%s)", lhs, quoteNumber(ik.AdjustmentValue))
								}

								rhs := quoteNumber(ik.Value)

								ruleTest := fmt.Sprintf("ok && (%s %s %s)", lhs, operator, rhs)
								emit("m%d = %s", rule.Level, ruleTest)
								emit("ml = %d", ik.ByteWidth)
							})
							emit("}")

						}
					case wizparser.KindFamilyString:
						sk, _ := rule.Kind.Data.(*wizparser.StringKind)
						emit("ml = i64(wizardry.StringTest(tb, int(off), %#v, %#v))", sk.Value, sk.Flags)
						if sk.Negate {
							emit("m%d = ml < 0", rule.Level)
						} else {
							emit("m%d = ml >= 0", rule.Level)
						}

					case wizparser.KindFamilySearch:
						sk, _ := rule.Kind.Data.(*wizparser.SearchKind)
						emit("ml = i64(wizardry.SearchTest(tb, int(off), %s, %s))", quoteNumber(int64(sk.MaxLen)), strconv.Quote(string(sk.Value)))
						// little trick for gof to be updated correctly
						emit("if ml >= 0 { ml += %s; m%d = true }",
							quoteNumber(int64(len(sk.Value))), rule.Level)

					default:
						emit("// uh oh unhandled kind")
						emitted = false
					}

					if emitted {
						emit("if m%d {", rule.Level)
						withIndent(func() {
							emit("fmt.Printf(\"matched rule: %%s\\n\", %s)", strconv.Quote(rule.Line))
							emit("gof = off + ml")
							if len(rule.Description) > 0 {
								emit("out = append(out, %s)", strconv.Quote(string(rule.Description)))
							}
						})
						emit("}")
					}

					if needsBreak {
						emit("break rule%d", ruleIndex)
						outdent()
						emit("}")
						outdent()
					}

					emit("")
				}

				for currentLevel > 0 {
					currentLevel--
					outdent()
					emit("}")
				}

				emit("return out, nil")
			})
			emit("}")
			emit("")
		}

	}
	return nil
}

func pageSymbol(page string, swapEndian bool) string {
	result := ""
	for _, token := range strings.Split(page, "-") {
		result += strings.Title(token)
	}

	if swapEndian {
		result += "__Swapped"
	}

	return result
}

func endiannessString(en wizparser.Endianness, swapEndian bool) string {
	if en.MaybeSwapped(swapEndian) == wizparser.BigEndian {
		return "be"
	}
	return "le"
}

func quoteNumber(number int64) string {
	if number < 0 {
		return fmt.Sprintf("%d", number)
	}
	return fmt.Sprintf("0x%x", number)
}
