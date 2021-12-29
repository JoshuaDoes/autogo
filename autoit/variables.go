package autoit

type VarType int
const (
	VarTypeInteger VarType = iota + 1
	VarTypeBoolean
	VarTypeString
	VarTypeFloat
	VarTypeArray
	VarTypeMap
	VarTypeBinary
	VarTypePointer
	VarTypeObject
	VarTypeStruct
	VarTypeFunction
)
type Var struct {
	varType VarType
	data interface{}
}