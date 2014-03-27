{{if ne .hintJQueryAlreadyLoaded "true"}}
CensorshipMeter.loadJQuery = function() {
  var headTag = document.getElementsByTagName('head')[0];
  var jqTag = document.createElement('script');
  jqTag.type = 'text/javascript';
  jqTag.src = '{{.serverUrl}}/jquery.js';
  jqTag.onload = function() {
    CensorshipMeter.run();
  }
  headTag.appendChild(jqTag);
}
{{end}}

{{if eq .hintJQueryAlreadyLoaded "true"}}
CensorshipMeter.run();
{{else if eq .hintJQueryAlreadyLoaded "false"}}
CensorshipMeter.loadJQuery();
{{else}}
if (typeof jQuery == "undefined") {
  CensorshipMeter.loadJQuery();
} else {
  CensorshipMeter.run();
}
{{end}}
