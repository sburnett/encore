var CensorshipMeter = new Object();
CensorshipMeter.baseUrl = "{{.serverUrl}}/submit";
CensorshipMeter.measurementId = encodeURIComponent("{{.measurementId}}");
CensorshipMeter.submitResult = function(state) {
  this.submitted = state;
  $.ajax({
    url: this.baseUrl + "?cmh-id=" + this.measurementId + "&cmh-result=" + encodeURIComponent(state),
  });
}
CensorshipMeter.sendSuccess = function() {
  this.submitResult("success");
}
CensorshipMeter.sendFailure = function() {
  this.submitResult("failure");
}
CensorshipMeter.sendException = function() {
  this.submitResult("exception");
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
    CensorshipMeter.measure();
  });
{{if ne .hintShowStats "false"}}
  $(function() {
    CensorshipMeter.setupStats();
  });
{{end}}
}
