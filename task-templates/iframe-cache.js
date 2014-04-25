{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var iframe = $('<iframe />');
  iframe.attr('width', 0);
  iframe.attr('height', 0);
  iframe.attr('src', '{{.iframeUrl}}');
  iframe.css('display', 'none');
  iframe.on('load', function() {
    try {
      var img = $('<img />');
      img.css('display', 'none');
      img.attr('src', '{{.imageUrl}}');
      img.on('load', function() {
        try {
          var endTime = $.now();
          CensorshipMeter.submitResult("load-time", endTime - CensorshipMeter.startTime);
        } catch(err) {
          CensorshipMeter.sendException(err);
        }
      });
      img.on('error', function() {
        CensorshipMeter.sendError();
      });
      CensorshipMeter.startTime = $.now();
      img.appendTo('html');
    } catch(err) {
      CensorshipMeter.sendException(err);
    }
  });
  iframe.appendTo('html');
}
{{template "footer.js" .}}
