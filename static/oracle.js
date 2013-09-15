// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

var oracle = (function() {

'use strict';

var modes = [
  {id: 'describe', name: 'Describe', desc: 'Describe the expression at the current point.'},
  {id: 'callees', name: 'Call targets', desc: 'Show possible callees of the function call at the current point.'},
  {id: 'callers', name: 'Callers', desc: 'Show the set of callers of the function containing the current point.'},
  {id: 'callgraph', name: 'Call graph', desc: 'Show the callgraph of the current program.'},
  {id: 'callstack', name: 'Call stack', desc: 'Show an arbitrary path from a root of the call graph to the function containing the current point.'},
  {id: 'freevars', name: 'Free variables', desc: 'Enumerate the free variables of the current selection.'},
  {id: 'implements', name: 'Implements', desc: 'Describe the \'implements\' relation for types in the package containing the current point.'},
  {id: 'peers', name: 'Channel peers', desc: 'Enumerate the set of possible corresponding sends/receives for this channel receive/send operation.'},
  {id: 'referrers', name: 'Referrers', desc: 'Enumerate all references to the object denoted by the selected identifier.'}
];

var message = {
  wait: 'Consulting the oracle ...',
  error: 'An error occurred.'
};

function oraclify(code, out, file) {
  var menu = modeMenu();
  $('body').append(menu);
  hideOnEscape(menu, out);
  hideOnClickOff(menu, code);
  code.mouseup(function(e) {
    var sel = selectedRange();
    if (!isRangeWithinElement(sel, code)) {
      return;
    }
    menu.unbind('select').on('select', function(e, mode) {
      menu.hide();
      changeContent(out, message.wait);
      // FIXME: these are character offsets, the oracle wants byte offsets
      query(mode, pos(file, sel.startOffset, sel.endOffset), 'plain')
        .done(function(data) {
          onResult(out, data);
        })
        .fail(function(e) {
          changeContent(out, message.error);
        });
    });
    menu.css({top: e.pageY, left: e.pageX});
    menu.show();
  });
}

var ESC = 27;

function hideOnEscape(menu, out) {
  $(document).keyup(function(e) {
    if (e.keyCode == ESC) {
      if (menu.is(':visible')) {
        menu.hide();
        return;
      }
      out.trigger('esc');
    }
  });
}

function hideOnClickOff(menu, code) {
  $('body').click(function(e) {
    if (!$(e.target).closest(code).length) {
      menu.hide();
    }
  });
}

function onResult(out, data) {
  changeContent(out, data);
}

function changeContent(element, text) {
  appendLinkified(element.empty(), text).trigger('change');
}

// file:line.col-line.col:
var rangeAddress = /(.*):([0-9]+)\.([0-9]+)-([0-9]+)\.([0-9]+): (.*)/;
// file:line:col:
var singleAddress = /(.*):([0-9]+):([0-9]+): (.*)/;
// -:
var noAddress = /-: (.*)/;

function appendLinkified(element, text) {
  var match, arrow = '▶ ';
  var lines = text.split('\n');
  var n = lines.length;
  for (var i = 0; i < n; i++) {
    var line = lines[i];
    if (match = rangeAddress.exec(line)) {
      var file = match[1];
      var fromLine = match[2];
      var fromCol = match[3];
      var toLine = match[4];
      var toCol = match[5];
      var rest = match[6];
      var link = sourceLink(file, fromLine, arrow + rest);
      element.append(link).append('\n');
      continue;
    }
    if (match = singleAddress.exec(line)) {
      var file = match[1];
      var line = match[2];
      var col = match[3];
      var rest = match[4];
      var link = sourceLink(file, line, arrow + rest);
      element.append(link).append('\n');
      continue;
    }
    if (match = noAddress.exec(line)) {
      var rest = match[1];
      element.append('  ' + rest + '\n');
      continue;
    }
    element.append(line + '\n');
  }
  return element;
}

function sourceLink(file, line, text) {
  return $('<a>').attr('href', sourceURL(file, line)).text(text);
}

function sourceURL(file, line) {
  return 'source?' + $.param({'file': file}) + '#L' + line;
}

function selectedRange() {
  return window.getSelection().getRangeAt(0);
}

function isRangeWithinElement(range, elem) {
  return range.commonAncestorContainer.parentElement == elem[0];
}

function query(mode, pos, format) {
  var data = {
    mode: mode,
    pos: pos,
    format: format
  };
  var get = (format == 'json') ? $.getJSON : $.get;
  return get('query', data);
}

function pos(file, start, end) {
  var p = file + ':#' + start;
  if (start != end) {
    p += ',#' + end;
  }
  return p;
}

function modeMenu() {
  var m = $('<ul class="menu">').hide();
  $.each(modes, function(i, mode) {
    var item = $('<li>').text(mode.name).attr('title', mode.desc);
    item.click(function() {
      m.trigger('select', mode.id);
    });
    m.append(item);
  });
  return m;
}

return {
  oraclify: oraclify
};

})();
