var CensorshipMeter = new Object();
CensorshipMeter.baseUrl = "{{.serverUrl}}/submit";
CensorshipMeter.measurementId = encodeURIComponent("{{.measurementId}}");
CensorshipMeter.maxMessageLength = 64;
CensorshipMeter.submitResult = function(state, message) {
  this.submitted = state;
  if (message != null) {
    message = String(message).substring(0, this.maxMessageLength);
  }
  var params = {
    "cmh-id": this.measurementId,
    "cmh-result": state,
    "cmh-message": message,
  };
  $.ajax({
    url: this.baseUrl + "?" + $.param(params),
  });
}
CensorshipMeter.sendSuccess = function() {
  this.submitResult("success");
}
CensorshipMeter.sendFailure = function() {
  this.submitResult("failure");
}
CensorshipMeter.sendException = function(err) {
  this.submitResult("exception", err);
}
{{if ne .hintShowStats "false"}}
CensorshipMeter.setupStats = function() {
  this.logo = $("#encore-stats");
  if (typeof this.logo == "undefined") {
    return;
  }
  {{if eq .count "0"}}
  this.logo.html('Visitors of this page automatically measure Web filtering. <a href="{{.serverUrl}}/stats/refer">Learn more</a>.');
  {{else if eq .count "1"}}
  this.logo.html('Visitors of this page have performed {{.count}} measurement of Web filtering. <a href="{{.serverUrl}}/stats/refer">Learn more</a>.');
  {{else}}
  this.logo.html('Visitors of this page have performed {{.count}} measurements of Web filtering. <a href="{{.serverUrl}}/stats/refer">Learn more</a>.');
  {{end}}
}
{{end}}
CensorshipMeter.run = function() {
  this.submitResult("init");
  $(function() {
    try {
      CensorshipMeter.measure();
    } catch(err) {
      CensorshipMeter.sendException(err);
    }
  });
{{if ne .hintShowStats "false"}}
  $(function() {
    CensorshipMeter.setupStats();
  });
{{end}}
}
