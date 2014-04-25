{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var iframe = $('<iframe />');
  iframe.attr('width', 0);
  iframe.attr('height', 0);
  iframe.attr('src', '{{.iframeUrl}}');
  iframe.css('display', 'none');
  iframe.on('load', function() {
    try {
      var endTime = $.now();
      CensorshipMeter.submitResult("load-time", endTime - CensorshipMeter.startTime);
    } catch(err) {
      CensorshipMeter.sendException(err);
    }
  });
  CensorshipMeter.startTime = $.now();
  iframe.appendTo('html');
}
{{template "footer.js" .}}
