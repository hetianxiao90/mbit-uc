package captcha

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
	"log"
	"strconv"
	"strings"
)

var slideBasicCapt slide.Captcha

type SlideBasicCaptData struct {
	ImageBase64 string
	TileBase64  string
	BlockData   *slide.Block
}

type CheckSlideData struct {
	Point         string
	Key           string
	CacheDataByte []byte
}

func init() {
	builder := slide.NewBuilder(
		//slide.WithGenGraphNumber(2),
		slide.WithEnableGraphVerticalRandom(true),
	)

	// background images
	imgs, err := images.GetImages()
	if err != nil {
		log.Fatalln(err)
	}

	graphs, err := tiles.GetTiles()
	if err != nil {
		log.Fatalln(err)
	}

	var newGraphs = make([]*slide.GraphImage, 0, len(graphs))
	for i := 0; i < len(graphs); i++ {
		graph := graphs[i]
		newGraphs = append(newGraphs, &slide.GraphImage{
			OverlayImage: graph.OverlayImage,
			MaskImage:    graph.MaskImage,
			ShadowImage:  graph.ShadowImage,
		})
	}

	// set resources
	builder.SetResources(
		slide.WithGraphImages(newGraphs),
		slide.WithBackgrounds(imgs),
	)

	slideBasicCapt = builder.Make()
}

func GetSlideBasic() (error, *SlideBasicCaptData) {
	captData, err := slideBasicCapt.Generate()
	if err != nil {
		log.Fatalln(err)
	}

	blockData := captData.GetData()
	if blockData == nil {
		return errors.New("gen captcha data failed"), nil
	}

	var masterImageBase64, tileImageBase64 string
	masterImageBase64 = captData.GetMasterImage().ToBase64()
	if err != nil {
		return errors.New("masterImageBase64 base64 data failed"), nil
	}

	tileImageBase64 = captData.GetTileImage().ToBase64()
	if err != nil {
		return errors.New("tileImageBase64 base64 data failed"), nil
	}

	return nil, &SlideBasicCaptData{
		ImageBase64: masterImageBase64,
		TileBase64:  tileImageBase64,
		BlockData:   blockData,
	}
}

func CheckSlide(d *CheckSlideData) error {

	point := d.Point
	key := d.Key
	if point == "" || key == "" {
		return errors.New("point or key param is empty")
	}
	cacheDataByte := d.CacheDataByte
	if len(cacheDataByte) == 0 {
		return errors.New("illegal key")
	}
	src := strings.Split(point, ",")

	var dct *slide.Block
	if err := json.Unmarshal(cacheDataByte, &dct); err != nil {
		return errors.New("illegal key")
	}

	chkRet := false
	if 2 == len(src) {
		sx, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[0]), 64)
		sy, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[1]), 64)
		chkRet = slide.CheckPoint(int64(sx), int64(sy), int64(dct.X), int64(dct.Y), 4)
	}

	if !chkRet {
		return errors.New("CheckPoint false")
	}
	return nil
}
