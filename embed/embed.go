package emded

import (
	babel "github/efjoubert/lnksys/babel"
	babylon "github/efjoubert/lnksys/babylon"
	bootstrap "github/efjoubert/lnksys/bootstrap"
	fontawesome "github/efjoubert/lnksys/fontawesome"
	jquery "github/efjoubert/lnksys/jquery"
	jss "github/efjoubert/lnksys/jss"
	materialdb "github/efjoubert/lnksys/materialdb"
	react "github/efjoubert/lnksys/react"
	require "github/efjoubert/lnksys/require"
	rxjs "github/efjoubert/lnksys/rxjs"
	three "github/efjoubert/lnksys/three"
	falcor "github/efjoubert/lnksys/falcor"
	video "github/efjoubert/lnksys/video"
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
