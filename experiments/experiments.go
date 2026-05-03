// TODO:
// Add property Mark int (public)
// Add method SetMark() int
// Add method GetMark() *Instruction
//
// BUT: I should then INSERT elements
// bcause with current implementation its just for read
// SO:
// I can try implementing Cursor instead
// and keep track of current program state
//
// BTW:
// Do we really need mark functionality?

package experiments

import "bleeder/internal/bleeder"

func Run() {
	content := `
	:c3 4_ |+7
	:e4 _4 >60 4 _+4 | +2
	`
	bleeder.ParseRaw(content, 0)

}
