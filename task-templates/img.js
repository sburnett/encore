{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var img = $('<img />');
  img.attr('src', '{{.imageUrl}}');
  img.css('display', 'none');
  img.on('load', function() {
    CensorshipMeter.sendSuccess();
  });
  img.on('error', function() {
    CensorshipMeter.sendFailure();
  });
  img.appendTo('html');
}
{{template "footer.js" .}}
