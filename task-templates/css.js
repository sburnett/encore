{{template "header.js" .}}
CensorshipMeter.measure = function() {
  var expRef = $('<span id="{{.cssId}}"></span>');
  expRef.appendTo('html');

  {{if .controlCssId}}
  var controlRef = $('<span id="{{.controlCssId}}"></span>');
  controlRef.appendTo('html');
  {{end}}

  $.getScript('{{.serverUrl}}/lazyload.js', function() {
    LazyLoad.css('{{.cssUrl}}', function() {
      try {
        var cssTag = $('#{{.cssId}}');
        var style = window.getComputedStyle(cssTag[0]);
        var positionStyle = style.getPropertyValue('{{.cssAttribute}}');
        if (positionStyle == '{{.cssDesiredValue}}') {
          CensorshipMeter.sendSuccess();
        } else {
          CensorshipMeter.sendFailure();
        }

        {{if .controlCssId}}
        var controlCssTag = $('#{{.controlCssId}}');
        var controlStyle = window.getComputedStyle(controlCssTag[0]);
        var controlPositionStyle = controlStyle.getPropertyValue('{{.cssAttribute}}');
        if (controlPositionStyle != '{{.cssDesiredValue}}') {
          CensorshipMeter.submitResult('success-control');
        } else {
          CensorshipMeter.submitResult('failure-control');
        }
        {{end}}
      } catch(err) {
        CensorshipMeter.sendException(err);
      }
    });
  });
}
{{template "footer.js" .}}
