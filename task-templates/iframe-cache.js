{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var iframe = $('<iframe />');
  iframe.attr('width', 0);
  iframe.attr('height', 0);
  iframe.attr('src', '{{.iframeUrl}}');
  iframe.css('display', 'none');
  iframe.on('load', function() {
    try {
      var iframeEndTime = $.now();
      CensorshipMeter.submitResult('load-time-iframe', iframeEndTime - CensorshipMeter.iframeStartTime);

      var img = $('<img />');
      img.css('display', 'none');
      img.attr('src', '{{.imageUrl}}');
      img.on('load', function() {
        try {
          var imgEndTime = $.now();
          CensorshipMeter.submitResult('load-time-img', imgEndTime - CensorshipMeter.imgStartTime);
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
