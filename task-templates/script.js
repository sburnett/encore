{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var script = $('<script></script>');
  script.attr('src', '{{.scriptUrl}}');
  script.on('load', function() {
    CensorshipMeter.sendSuccess();
  });
  script.on('error', function() {
    CensorshipMeter.sendFailure();
  });
  script.appendTo('html');
}
{{template "footer.js" .}}
