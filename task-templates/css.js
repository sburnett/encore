{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var iframe = $('<iframe />');
  iframe.attr('width', 0);
  iframe.attr('height', 0);
  iframe.attr('srcdoc', '<link href="{{.cssUrl}}" rel="stylesheet" type="text/css"><p id="testParagraph"></p>');
  iframe.css('display', 'none');
  iframe.on('load', function() {
    try {
      var el = $(this);
      var p = $(this).contents().find('#testParagraph')[0];
      var style = window.getComputedStyle(p);
      var positionStyle = style.getPropertyValue('position');
      if (positionStyle == 'absolute') {
        CensorshipMeter.sendSuccess();
      } else {
        CensorshipMeter.sendFailure();
      }
    } catch(err) {
      CensorshipMeter.sendException();
    }
  });
  iframe.appendTo('html');
}
{{template "footer.js" .}}
