// Code generated by 'yaegi extract math'. DO NOT EDIT.

// +build go1.15,!go1.16

package stdlib

import (
	"go/constant"
	"go/token"
	"math"
	"reflect"
)

func init() {
	Symbols["math/math"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Abs":                    reflect.ValueOf(math.Abs),
		"Acos":                   reflect.ValueOf(math.Acos),
		"Acosh":                  reflect.ValueOf(math.Acosh),
		"Asin":                   reflect.ValueOf(math.Asin),
		"Asinh":                  reflect.ValueOf(math.Asinh),
		"Atan":                   reflect.ValueOf(math.Atan),
		"Atan2":                  reflect.ValueOf(math.Atan2),
		"Atanh":                  reflect.ValueOf(math.Atanh),
		"Cbrt":                   reflect.ValueOf(math.Cbrt),
		"Ceil":                   reflect.ValueOf(math.Ceil),
		"Copysign":               reflect.ValueOf(math.Copysign),
		"Cos":                    reflect.ValueOf(math.Cos),
		"Cosh":                   reflect.ValueOf(math.Cosh),
		"Dim":                    reflect.ValueOf(math.Dim),
		"E":                      reflect.ValueOf(constant.MakeFromLiteral("2.71828182845904523536028747135266249775724709369995957496696762566337824315673231520670375558666729784504486779277967997696994772644702281675346915668215131895555530285035761295375777990557253360748291015625", token.FLOAT, 0)),
		"Erf":                    reflect.ValueOf(math.Erf),
		"Erfc":                   reflect.ValueOf(math.Erfc),
		"Erfcinv":                reflect.ValueOf(math.Erfcinv),
		"Erfinv":                 reflect.ValueOf(math.Erfinv),
		"Exp":                    reflect.ValueOf(math.Exp),
		"Exp2":                   reflect.ValueOf(math.Exp2),
		"Expm1":                  reflect.ValueOf(math.Expm1),
		"FMA":                    reflect.ValueOf(math.FMA),
		"Float32bits":            reflect.ValueOf(math.Float32bits),
		"Float32frombits":        reflect.ValueOf(math.Float32frombits),
		"Float64bits":            reflect.ValueOf(math.Float64bits),
		"Float64frombits":        reflect.ValueOf(math.Float64frombits),
		"Floor":                  reflect.ValueOf(math.Floor),
		"Frexp":                  reflect.ValueOf(math.Frexp),
		"Gamma":                  reflect.ValueOf(math.Gamma),
		"Hypot":                  reflect.ValueOf(math.Hypot),
		"Ilogb":                  reflect.ValueOf(math.Ilogb),
		"Inf":                    reflect.ValueOf(math.Inf),
		"IsInf":                  reflect.ValueOf(math.IsInf),
		"IsNaN":                  reflect.ValueOf(math.IsNaN),
		"J0":                     reflect.ValueOf(math.J0),
		"J1":                     reflect.ValueOf(math.J1),
		"Jn":                     reflect.ValueOf(math.Jn),
		"Ldexp":                  reflect.ValueOf(math.Ldexp),
		"Lgamma":                 reflect.ValueOf(math.Lgamma),
		"Ln10":                   reflect.ValueOf(constant.MakeFromLiteral("2.30258509299404568401799145468436420760110148862877297603332784146804725494827975466552490443295866962642372461496758838959542646932914211937012833592062802600362869664962772731087170541286468505859375", token.FLOAT, 0)),
		"Ln2":                    reflect.ValueOf(constant.MakeFromLiteral("0.6931471805599453094172321214581765680755001343602552541206800092715999496201383079363438206637927920954189307729314303884387720696314608777673678644642390655170150035209453154294578780536539852619171142578125", token.FLOAT, 0)),
		"Log":                    reflect.ValueOf(math.Log),
		"Log10":                  reflect.ValueOf(math.Log10),
		"Log10E":                 reflect.ValueOf(constant.MakeFromLiteral("0.43429448190325182765112891891660508229439700580366656611445378416636798190620320263064286300825210972160277489744884502676719847561509639618196799746596688688378591625127711495224502868950366973876953125", token.FLOAT, 0)),
		"Log1p":                  reflect.ValueOf(math.Log1p),
		"Log2":                   reflect.ValueOf(math.Log2),
		"Log2E":                  reflect.ValueOf(constant.MakeFromLiteral("1.44269504088896340735992468100189213742664595415298593413544940772066427768997545329060870636212628972710992130324953463427359402479619301286929040235571747101382214539290471666532766903401352465152740478515625", token.FLOAT, 0)),
		"Logb":                   reflect.ValueOf(math.Logb),
		"Max":                    reflect.ValueOf(math.Max),
		"MaxFloat32":             reflect.ValueOf(constant.MakeFromLiteral("340282346638528859811704183484516925440", token.FLOAT, 0)),
		"MaxFloat64":             reflect.ValueOf(constant.MakeFromLiteral("179769313486231570814527423731704356798100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", token.FLOAT, 0)),
		"MaxInt16":               reflect.ValueOf(constant.MakeFromLiteral("32767", token.INT, 0)),
		"MaxInt32":               reflect.ValueOf(constant.MakeFromLiteral("2147483647", token.INT, 0)),
		"MaxInt64":               reflect.ValueOf(constant.MakeFromLiteral("9223372036854775807", token.INT, 0)),
		"MaxInt8":                reflect.ValueOf(constant.MakeFromLiteral("127", token.INT, 0)),
		"MaxUint16":              reflect.ValueOf(constant.MakeFromLiteral("65535", token.INT, 0)),
		"MaxUint32":              reflect.ValueOf(constant.MakeFromLiteral("4294967295", token.INT, 0)),
		"MaxUint64":              reflect.ValueOf(constant.MakeFromLiteral("18446744073709551615", token.INT, 0)),
		"MaxUint8":               reflect.ValueOf(constant.MakeFromLiteral("255", token.INT, 0)),
		"Min":                    reflect.ValueOf(math.Min),
		"MinInt16":               reflect.ValueOf(constant.MakeFromLiteral("-32768", token.INT, 0)),
		"MinInt32":               reflect.ValueOf(constant.MakeFromLiteral("-2147483648", token.INT, 0)),
		"MinInt64":               reflect.ValueOf(constant.MakeFromLiteral("-9223372036854775808", token.INT, 0)),
		"MinInt8":                reflect.ValueOf(constant.MakeFromLiteral("-128", token.INT, 0)),
		"Mod":                    reflect.ValueOf(math.Mod),
		"Modf":                   reflect.ValueOf(math.Modf),
		"NaN":                    reflect.ValueOf(math.NaN),
		"Nextafter":              reflect.ValueOf(math.Nextafter),
		"Nextafter32":            reflect.ValueOf(math.Nextafter32),
		"Phi":                    reflect.ValueOf(constant.MakeFromLiteral("1.6180339887498948482045868343656381177203091798057628621354486119746080982153796619881086049305501566952211682590824739205931370737029882996587050475921915678674035433959321750307935872115194797515869140625", token.FLOAT, 0)),
		"Pi":                     reflect.ValueOf(constant.MakeFromLiteral("3.141592653589793238462643383279502884197169399375105820974944594789982923695635954704435713335896673485663389728754819466702315787113662862838515639906529162340867271374644786874341662041842937469482421875", token.FLOAT, 0)),
		"Pow":                    reflect.ValueOf(math.Pow),
		"Pow10":                  reflect.ValueOf(math.Pow10),
		"Remainder":              reflect.ValueOf(math.Remainder),
		"Round":                  reflect.ValueOf(math.Round),
		"RoundToEven":            reflect.ValueOf(math.RoundToEven),
		"Signbit":                reflect.ValueOf(math.Signbit),
		"Sin":                    reflect.ValueOf(math.Sin),
		"Sincos":                 reflect.ValueOf(math.Sincos),
		"Sinh":                   reflect.ValueOf(math.Sinh),
		"SmallestNonzeroFloat32": reflect.ValueOf(constant.MakeFromLiteral("1.40129846432481707092372958328991613128000000000000000000000000000000000000000000001246655487714533538006789189734126694785975183981128816138510360971472225738624150874949653910667523779981133927289771669016713539217953030564201688027906006008453304556102801950542906382507e-45", token.FLOAT, 0)),
		"SmallestNonzeroFloat64": reflect.ValueOf(constant.MakeFromLiteral("4.94065645841246544176568792868221372365099999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999916206614696136086629714037163874026187912451674985660337336755242863513549746484310667379088263176934591818322489862214324814281481943599945502119376688748731948897748561110123901991443297110206447991752071007926740839424145013355231935665542622515363894390826799291671723318261174778903704064716351336223785714389641180220184242018383103204287325861250404139399888498504162666394779407509786431980433771341978183418568838015304951087487907666317075235615216699116844779095660202193409146032665221882798856203896125090454090026556150624798681464913851491093798848436664885581161128190046248588053014958829424991704801027040654863867512297941601850496672190315253109308532379657238854928816482120688440415705411555019932096150435627305446214567713171657554140575630917301482608119551500514805985376055777894871863446222606532650275466165274006e-324", token.FLOAT, 0)),
		"Sqrt":                   reflect.ValueOf(math.Sqrt),
		"Sqrt2":                  reflect.ValueOf(constant.MakeFromLiteral("1.414213562373095048801688724209698078569671875376948073176679739576083351575381440094441524123797447886801949755143139115339040409162552642832693297721230919563348109313505318596071447245776653289794921875", token.FLOAT, 0)),
		"SqrtE":                  reflect.ValueOf(constant.MakeFromLiteral("1.64872127070012814684865078781416357165377610071014801157507931167328763229187870850146925823776361770041160388013884200789716007979526823569827080974091691342077871211546646890155898290686309337615966796875", token.FLOAT, 0)),
		"SqrtPhi":                reflect.ValueOf(constant.MakeFromLiteral("1.2720196495140689642524224617374914917156080418400962486166403754616080542166459302584536396369727769747312116100875915825863540562126478288118732191412003988041797518382391984914647764526307582855224609375", token.FLOAT, 0)),
		"SqrtPi":                 reflect.ValueOf(constant.MakeFromLiteral("1.772453850905516027298167483341145182797549456122387128213807789740599698370237052541269446184448945647349951047154197675245574635259260134350885938555625028620527962319730619356050738133490085601806640625", token.FLOAT, 0)),
		"Tan":                    reflect.ValueOf(math.Tan),
		"Tanh":                   reflect.ValueOf(math.Tanh),
		"Trunc":                  reflect.ValueOf(math.Trunc),
		"Y0":                     reflect.ValueOf(math.Y0),
		"Y1":                     reflect.ValueOf(math.Y1),
		"Yn":                     reflect.ValueOf(math.Yn),
	}
}
