{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var img = $('<img>');
  img.attr('src', '{{.imageUrl}}');
  img.attr('style', 'display: none');
  img.attr('onload', 'CensorshipMeter.sendSuccess()');
  img.attr('onerror', 'CensorshipMeter.sendFailure()');
  img.appendTo('html');
}
{{template "footer.js" .}}
