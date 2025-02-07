package golibmagic

import "github.com/postfix/golibmagic/util"

// SearchTest looks for a fixed pattern at any position within a certain length
func SearchTest(sr *util.SliceReader, targetIndex int64, maxLen int64, pattern string) int64 {
	sf := MakeStringFinder(pattern)

	sr = sr.Slice(targetIndex).Cap(maxLen)
	return sf.next(sr)
}
