package babylon

import (
	"io"
	"strings"
)

func BabylonFindJS(babylonfindjs string) io.Reader {
	if strings.LastIndex(babylonfindjs, "/") >= 0 {
		babylonfindjs = babylonfindjs[strings.LastIndex(babylonfindjs, "/")+1:]
	}
	if babylonfindjs == "babylon.js" {
		return BabylonJS()
	} else if babylonfindjs == "babylon.inspector.bundle.js" {
		return BabylonInpectorJS()
	} else if babylonfindjs == "babylonjs.loaders.min.js" {
		return BabylonLoadersJS()
	} else if babylonfindjs == "babylonjs.materials.min.js" {
		return BabylonMaterialsJS()
	} else if babylonfindjs == "babylonjs.postProcess.min.js" {
		return BabylonPostProcessJS()
	} else if babylonfindjs == "babylonjs.proceduralTextures.min.js" {
		return BabylonProceduralTexturesJS()
	} else if babylonfindjs == "babylonjs.serializers.min.js" {
		return BabylonSerializersJS()
	} else if babylonfindjs == "ammo.js" {
		return AmmoJS()
	} else if babylonfindjs == "oimo.js" {
		return OimoJS()
	}
	return nil
}
