{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var iframe = $('<iframe />');
  iframe.attr('width', 0);
  iframe.attr('height', 0);
  iframe.attr('src', '{{.iframeUrl}}');
  iframe.css('display', 'none');
  iframe.on('load', function() {
    try {
      CensorshipMeter.iframeEndTime = $.now();
      var img = $('<img />');
      img.css('display', 'none');
      img.attr('src', '{{.imageUrl}}');
      img.on('load', function() {
        try {
          var imgEndTime = $.now();
          var iframeTime = CensorshipMeter.iframeEndTime - CensorshipMeter.iframeStartTime;
          var imgTime = imgEndTime - CensorshipMeter.imgStartTime;
          var message = iframeTime + ',' + imgTime;
          CensorshipMeter.submitResult('load-time', message);
        } catch(err) {
          CensorshipMeter.sendException(err);
        }
      });
      img.on('error', function() {
        CensorshipMeter.sendError();
      });
      CensorshipMeter.imgStartTime = $.now();
      img.appendTo('html');
    } catch(err) {
      CensorshipMeter.sendException(err);
    }
  });
  CensorshipMeter.iframeStartTime = $.now();
  iframe.appendTo('html');
}
{{template "footer.js" .}}
