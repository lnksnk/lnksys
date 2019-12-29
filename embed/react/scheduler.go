package react

import (
	"io"
	"strings"
)

const schedulertracingjs string = `/**
* @license React
*
* Copyright (c) Facebook, Inc. and its affiliates.
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/

'use strict';

(function(global, factory) {
 // eslint-disable-next-line no-unused-expressions
 typeof exports === 'object' && typeof module !== 'undefined'
   ? (module.exports = factory(require('react')))
   : typeof define === 'function' && define.amd // eslint-disable-line no-undef
	 ? define(['react'], factory) // eslint-disable-line no-undef
	 : (global.SchedulerTracing = factory(global));
})(this, function(global) {
 function unstable_clear() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_clear.apply(
	 this,
	 arguments
   );
 }

 function unstable_getCurrent() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_getCurrent.apply(
	 this,
	 arguments
   );
 }

 function unstable_getThreadID() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_getThreadID.apply(
	 this,
	 arguments
   );
 }

 function unstable_subscribe() {
   // eslint-disable-next-line max-len
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_subscribe.apply(
	 this,
	 arguments
   );
 }

 function unstable_trace() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_trace.apply(
	 this,
	 arguments
   );
 }

 function unstable_unsubscribe() {
   // eslint-disable-next-line max-len
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_unsubscribe.apply(
	 this,
	 arguments
   );
 }

 function unstable_wrap() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_wrap.apply(
	 this,
	 arguments
   );
 }

 return Object.freeze({
   unstable_clear: unstable_clear,
   unstable_getCurrent: unstable_getCurrent,
   unstable_getThreadID: unstable_getThreadID,
   unstable_subscribe: unstable_subscribe,
   unstable_trace: unstable_trace,
   unstable_unsubscribe: unstable_unsubscribe,
   unstable_wrap: unstable_wrap,
 });
});`

func SchedulerTracingJS() io.Reader {
	return strings.NewReader(schedulertracingjs)
}

const schedulertracingprofilingjs string = `/**
* @license React
*
* Copyright (c) Facebook, Inc. and its affiliates.
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/

'use strict';

(function(global, factory) {
 // eslint-disable-next-line no-unused-expressions
 typeof exports === 'object' && typeof module !== 'undefined'
   ? (module.exports = factory(require('react')))
   : typeof define === 'function' && define.amd // eslint-disable-line no-undef
	 ? define(['react'], factory) // eslint-disable-line no-undef
	 : (global.SchedulerTracing = factory(global));
})(this, function(global) {
 function unstable_clear() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_clear.apply(
	 this,
	 arguments
   );
 }

 function unstable_getCurrent() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_getCurrent.apply(
	 this,
	 arguments
   );
 }

 function unstable_getThreadID() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_getThreadID.apply(
	 this,
	 arguments
   );
 }

 function unstable_subscribe() {
   // eslint-disable-next-line max-len
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_subscribe.apply(
	 this,
	 arguments
   );
 }

 function unstable_trace() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_trace.apply(
	 this,
	 arguments
   );
 }

 function unstable_unsubscribe() {
   // eslint-disable-next-line max-len
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_unsubscribe.apply(
	 this,
	 arguments
   );
 }

 function unstable_wrap() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.SchedulerTracing.unstable_wrap.apply(
	 this,
	 arguments
   );
 }

 return Object.freeze({
   unstable_clear: unstable_clear,
   unstable_getCurrent: unstable_getCurrent,
   unstable_getThreadID: unstable_getThreadID,
   unstable_subscribe: unstable_subscribe,
   unstable_trace: unstable_trace,
   unstable_unsubscribe: unstable_unsubscribe,
   unstable_wrap: unstable_wrap,
 });
});`

func SchedulerTracingProfilingJS() io.Reader {
	return strings.NewReader(schedulertracingprofilingjs)
}

const schedulerjs string = `/**
* @license React
*
* Copyright (c) Facebook, Inc. and its affiliates.
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/

/* eslint-disable max-len */

'use strict';

(function(global, factory) {
 // eslint-disable-next-line no-unused-expressions
 typeof exports === 'object' && typeof module !== 'undefined'
   ? (module.exports = factory(require('react')))
   : typeof define === 'function' && define.amd // eslint-disable-line no-undef
	 ? define(['react'], factory) // eslint-disable-line no-undef
	 : (global.Scheduler = factory(global));
})(this, function(global) {
 function unstable_now() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_now.apply(
	 this,
	 arguments
   );
 }

 function unstable_scheduleCallback() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_scheduleCallback.apply(
	 this,
	 arguments
   );
 }

 function unstable_cancelCallback() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_cancelCallback.apply(
	 this,
	 arguments
   );
 }

 function unstable_shouldYield() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_shouldYield.apply(
	 this,
	 arguments
   );
 }

 function unstable_requestPaint() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_requestPaint.apply(
	 this,
	 arguments
   );
 }

 function unstable_runWithPriority() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_runWithPriority.apply(
	 this,
	 arguments
   );
 }

 function unstable_next() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_next.apply(
	 this,
	 arguments
   );
 }

 function unstable_wrapCallback() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_wrapCallback.apply(
	 this,
	 arguments
   );
 }

 function unstable_getCurrentPriorityLevel() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_getCurrentPriorityLevel.apply(
	 this,
	 arguments
   );
 }

 function unstable_getFirstCallbackNode() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_getFirstCallbackNode.apply(
	 this,
	 arguments
   );
 }

 function unstable_pauseExecution() {
   return undefined;
 }

 function unstable_continueExecution() {
   return undefined;
 }

 function unstable_forceFrameRate() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_forceFrameRate.apply(
	 this,
	 arguments
   );
 }

 return Object.freeze({
   unstable_now: unstable_now,
   unstable_scheduleCallback: unstable_scheduleCallback,
   unstable_cancelCallback: unstable_cancelCallback,
   unstable_shouldYield: unstable_shouldYield,
   unstable_requestPaint: unstable_requestPaint,
   unstable_runWithPriority: unstable_runWithPriority,
   unstable_next: unstable_next,
   unstable_wrapCallback: unstable_wrapCallback,
   unstable_getCurrentPriorityLevel: unstable_getCurrentPriorityLevel,
   unstable_continueExecution: unstable_continueExecution,
   unstable_pauseExecution: unstable_pauseExecution,
   unstable_getFirstCallbackNode: unstable_getFirstCallbackNode,
   unstable_forceFrameRate: unstable_forceFrameRate,
   get unstable_IdlePriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_IdlePriority;
   },
   get unstable_ImmediatePriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_ImmediatePriority;
   },
   get unstable_LowPriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_LowPriority;
   },
   get unstable_NormalPriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_NormalPriority;
   },
   get unstable_UserBlockingPriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_UserBlockingPriority;
   },
   get unstable_Profiling() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_Profiling;
   },
 });
});`

func SchedulerJS() io.Reader {
	return strings.NewReader(schedulerjs)
}

const schedulerprofilingjs string = `/**
* @license React
*
* Copyright (c) Facebook, Inc. and its affiliates.
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/

/* eslint-disable max-len */

'use strict';

(function(global, factory) {
 // eslint-disable-next-line no-unused-expressions
 typeof exports === 'object' && typeof module !== 'undefined'
   ? (module.exports = factory(require('react')))
   : typeof define === 'function' && define.amd // eslint-disable-line no-undef
	 ? define(['react'], factory) // eslint-disable-line no-undef
	 : (global.Scheduler = factory(global));
})(this, function(global) {
 function unstable_now() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_now.apply(
	 this,
	 arguments
   );
 }

 function unstable_scheduleCallback() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_scheduleCallback.apply(
	 this,
	 arguments
   );
 }

 function unstable_cancelCallback() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_cancelCallback.apply(
	 this,
	 arguments
   );
 }

 function unstable_shouldYield() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_shouldYield.apply(
	 this,
	 arguments
   );
 }

 function unstable_requestPaint() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_requestPaint.apply(
	 this,
	 arguments
   );
 }

 function unstable_runWithPriority() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_runWithPriority.apply(
	 this,
	 arguments
   );
 }

 function unstable_next() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_next.apply(
	 this,
	 arguments
   );
 }

 function unstable_wrapCallback() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_wrapCallback.apply(
	 this,
	 arguments
   );
 }

 function unstable_getCurrentPriorityLevel() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_getCurrentPriorityLevel.apply(
	 this,
	 arguments
   );
 }

 function unstable_getFirstCallbackNode() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_getFirstCallbackNode.apply(
	 this,
	 arguments
   );
 }

 function unstable_pauseExecution() {
   return undefined;
 }

 function unstable_continueExecution() {
   return undefined;
 }

 function unstable_forceFrameRate() {
   return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.Scheduler.unstable_forceFrameRate.apply(
	 this,
	 arguments
   );
 }

 return Object.freeze({
   unstable_now: unstable_now,
   unstable_scheduleCallback: unstable_scheduleCallback,
   unstable_cancelCallback: unstable_cancelCallback,
   unstable_shouldYield: unstable_shouldYield,
   unstable_requestPaint: unstable_requestPaint,
   unstable_runWithPriority: unstable_runWithPriority,
   unstable_next: unstable_next,
   unstable_wrapCallback: unstable_wrapCallback,
   unstable_getCurrentPriorityLevel: unstable_getCurrentPriorityLevel,
   unstable_continueExecution: unstable_continueExecution,
   unstable_pauseExecution: unstable_pauseExecution,
   unstable_getFirstCallbackNode: unstable_getFirstCallbackNode,
   unstable_forceFrameRate: unstable_forceFrameRate,
   get unstable_IdlePriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_IdlePriority;
   },
   get unstable_ImmediatePriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_ImmediatePriority;
   },
   get unstable_LowPriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_LowPriority;
   },
   get unstable_NormalPriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_NormalPriority;
   },
   get unstable_UserBlockingPriority() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_UserBlockingPriority;
   },
   get unstable_Profiling() {
	 return global.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED
	   .Scheduler.unstable_Profiling;
   },
 });
});`

func SchedulerProfilingJS() io.Reader {
	return strings.NewReader(schedulerprofilingjs)
}

func SchedulerFindJS(schedulerfindjs string) io.Reader {
	if strings.LastIndex(schedulerfindjs, "/") >= 0 {
		schedulerfindjs = schedulerfindjs[strings.LastIndex(schedulerfindjs, "/")+1:]
	}
	if schedulerfindjs == "scheduler.js" {
		return SchedulerJS()
	} else if schedulerfindjs == "scheduler.profiling.js" {
		return SchedulerProfilingJS()
	} else if schedulerfindjs == "scheduler-tracing.js" {
		return SchedulerTracingJS()
	} else if schedulerfindjs == "scheduler-tracing.profiling.js" {
		return SchedulerTracingProfilingJS()
	}
	return nil
}
