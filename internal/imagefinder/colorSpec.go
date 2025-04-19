package imagefinder

// ColorSpec описывает, какие пиксели считаем «попаданием».
type ColorSpec struct {
	HueRanges      [][2]float32 // список диапазонов H (0‑179)
	MinSat, MaxSat float32      // диапазон S (0‑1)
	MinVal, MaxVal float32      // диапазон V (0‑1)
}

var colorSpecs = map[string]ColorSpec{
	"red": {
		HueRanges: [][2]float32{{0, 10}, {170, 179}},
		MinSat:    0.4, MaxSat: 1,
		MinVal: 0.3, MaxVal: 1,
	},
	"green": {
		HueRanges: [][2]float32{{40, 80}},
		MinSat:    0.4, MaxSat: 1,
		MinVal: 0.3, MaxVal: 1,
	},
	"blue": {
		HueRanges: [][2]float32{{100, 140}},
		MinSat:    0.4, MaxSat: 1,
		MinVal: 0.6, MaxVal: 1,
	},
	"yellow": {
		HueRanges: [][2]float32{{15, 35}},
		MinSat:    0.4, MaxSat: 1,
		MinVal: 0.4, MaxVal: 1,
	},
	"gray": { // «серая»/задиммленная кнопка
		HueRanges: [][2]float32{{0, 179}}, // hue любой
		MinSat:    0, MaxSat: 0.25,        // почти нет насыщенности
		MinVal: 0.2, MaxVal: 0.9, // и не совсем чёрный/белый
	},
}
