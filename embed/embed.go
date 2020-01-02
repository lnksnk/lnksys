package emded

import (
	babel "github.com/efjoubert/lnksys/embed/babel"
	babylon "github.com/efjoubert/lnksys/embed/babylon"
	bootstrap "github.com/efjoubert/lnksys/embed/bootstrap"
	fontawesome "github.com/efjoubert/lnksys/embed/fontawesome"
	jquery "github.com/efjoubert/lnksys/embed/jquery"
	jss "github.com/efjoubert/lnksys/embed/jss"
	materialdb "github.com/efjoubert/lnksys/embed/materialdb"
	react "github.com/efjoubert/lnksys/embed/react"
	require "github.com/efjoubert/lnksys/embed/require"
	rxjs "github.com/efjoubert/lnksys/embed/rxjs"
	three "github.com/efjoubert/lnksys/embed/three"
	falcor "github.com/efjoubert/lnksys/embed/falcor"
	video "github.com/efjoubert/lnksys/embed/video"
	jspanel "github.com/efjoubert/lnksys/embed/jspanel"
	ace "github.com/efjoubert/lnksys/embed/ace"
	chart "github.com/efjoubert/lnksys/embed/chart"
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
	} else if embedjs=jspanel.JSPanelFindJSCSS(embedfindjs); embedjs!=nil {
		return
	} else if embedjs=ace.AcsJSFindJS(embedfindjs); embedjs!=nil {
		return
	} else if embedjs=chart.ChartFindJS(embedfindjs); embedjs!=nil {
		return
	}
	return
}
