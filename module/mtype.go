package module

// Type 组件类型
type Type string

const (
	// TYPE_DOWNLOADER 下载器
	TYPE_DOWNLOADER Type = "downloader"
	// TYPE_ANALYZER 分析器
	TYPE_ANALYZER Type = "analyzer"
	// TYPE_PIPELINE 条目处理管道
	TYPE_PIPELINE Type = "pipeline"
)

// legalTypeLetterMap 合法的组件类型-字母的映射
var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER:   "A",
	TYPE_PIPELINE:   "P",
}

// legalLetterTypeMap 合法的字母-组件类型的映射
var legalLetterTypeMap = map[string]Type{
	"D": TYPE_DOWNLOADER,
	"A": TYPE_ANALYZER,
	"P": TYPE_PIPELINE,
}

// CheckType 判断组件实例类型是否匹配
func CheckType(moduleType Type, module Module) bool {
	if moduleType == "" || module == nil {
		return false
	}
	switch moduleType {
	case TYPE_DOWNLOADER:
		if _, ok := module.(Downloader); ok {
			return true
		}
	case TYPE_ANALYZER:
		if _, ok := module.(Analyzer); ok {
			return true
		}
	case TYPE_PIPELINE:
		if _, ok := module.(Pipeline); ok {
			return true
		}
	}
	return false
}

// LegalType 判断给定组件类型是否合法
func LegalType(moduleType Type) bool {
	if _, ok := legalTypeLetterMap[moduleType]; ok {
		return true
	}
	return false
}

// GetType 获取组件类型
// 若给定组件ID不合法，则第一个返回结果值为false
func GetType(mid MID) (bool, Type) {
	parts, err := SplitMID(mid)
	if err != nil {
		return false, ""
	}
	mt, ok := legalLetterTypeMap[parts[0]]
	return ok, mt
}

// getLetter 获取组件类型的字母代号
func getLetter(moduleType Type) (found bool, letter string) {
	for l, t := range legalLetterTypeMap {
		if t == moduleType {
			letter = l
			found = true
			return
		}
	}
	return
}

// typeToLetter 根据组件类型获取字母代号
// 若给定组件类型不合法，则第一个结果返回false
func typeToLetter(moduleType Type) (bool, string) {
	switch moduleType {
	case TYPE_DOWNLOADER:
		return true, "D"
	case TYPE_ANALYZER:
		return true, "A"
	case TYPE_PIPELINE:
		return true, "P"
	default:
		return false, ""
	}
}

// letterToType 根据字母代号获取组件类型
// 若给定字母代号不合法，则第一个结果返回false
func letterToType(letter string) (bool, Type) {
	switch letter {
	case "D":
		return true, TYPE_DOWNLOADER
	case "A":
		return true, TYPE_ANALYZER
	case "P":
		return true, TYPE_PIPELINE
	default:
		return false, ""
	}
}
