package emded

import (
	babel "github.com/efjoubert/lnksys/babel"
	babylon "github.com/efjoubert/lnksys/babylon"
	bootstrap "github.com/efjoubert/lnksys/bootstrap"
	fontawesome "github.com/efjoubert/lnksys/fontawesome"
	jquery "github.com/efjoubert/lnksys/jquery"
	jss "github.com/efjoubert/lnksys/jss"
	materialdb "github.com/efjoubert/lnksys/materialdb"
	react "github.com/efjoubert/lnksys/react"
	require "github.com/efjoubert/lnksys/require"
	rxjs "github.com/efjoubert/lnksys/rxjs"
	three "github.com/efjoubert/lnksys/three"
	falcor "github.com/efjoubert/lnksys/falcor"
	video "github.com/efjoubert/lnksys/video"
	"io"
)

func EmbedFindJS(embedfindjs string) (embedjs io.Reader) {
	if embedjs = react.ReactFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = react.SchedulerFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = jquery.JQueryFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = require.RequireFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = bootstrap.BootstrapFindJSCSS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = fontawesome.FontawesomeFindJSCSS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = materialdb.MaterialDBFindJSCSS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = babel.BabelFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = babylon.BabylonFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = three.ThreeFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = jss.JSSFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = rxjs.RxJSFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = falcor.FalcorFindJS(embedfindjs); embedjs != nil {
		return
	} else if embedjs = video.VideoFindJSCSS(embedfindjs); embedjs != nil {
		return
	}
	return
}
