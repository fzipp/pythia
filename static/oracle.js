// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

var oracle = (function() {

'use strict';

var Filter = {
  YES: function(range) { return true; },
  NO: function(range) { return false; },
  RANGE: function(range) { return range.startOffset != range.endOffset; }
};

var modes = [
  {id: 'describe', menu: Filter.YES, name: 'Describe',
   desc: 'Describe the expression at the current point.'},
  {id: 'callees', menu: Filter.YES, name: 'Call targets',
   desc: 'Show possible callees of the function call at the current point.'},
  {id: 'callers', menu: Filter.YES, name: 'Callers',
   desc: 'Show the set of callers of the function containing the current point.'},
  {id: 'callgraph', menu: Filter.NO, name: 'Call graph',
   desc: 'Show the callgraph of the current program.'},
  {id: 'callstack', menu: Filter.YES, name: 'Call stack',
   desc: 'Show an arbitrary path from a root of the call graph to the function containing the current point.'},
  {id: 'freevars', menu: Filter.RANGE, name: 'Free variables',
   desc: 'Enumerate the free variables of the current typeection.'},
  {id: 'implements', menu: Filter.NO, name: 'Implements',
   desc: 'Describe the \'implements\' relation for types in the package containing the current point.'},
  {id: 'peers', menu: Filter.YES, name: 'Channel peers',
   desc: 'Enumerate the set of possible corresponding sends/receives for this channel receive/send operation.'},
  {id: 'referrers', menu: Filter.YES, name: 'Referrers',
   desc: 'Enumerate all references to the object denoted by the typeected identifier.'}
];

var message = {
  wait: 'Consulting the oracle ...',
  error: 'An error occurred.'
};

var title = 'Go source code oracle';

var currentFile;
var out, nums, code;

function init(source, output, file) {
  out = output;
  nums = source.find('.nums');
  code = source.find('.lines');
  currentFile = file;

  loadAndShowSource(file);

  var menu = modeMenu();
  $('body').append(menu);
  hideOnEscape(menu, out);
  hideOnClickOff(menu);
  code.mouseup(function(e) {
    var sel = selectedRange();
    if (!isRangeWithinElement(sel, code)) {
      return;
    }
    menu.unbind('select').on('select', function(e, mode) {
      menu.hide();
      var b = getByteOffsets(code.text(), sel);
      queryAction(mode, b.startOffset, b.endOffset);
    });
    filterModes(menu, sel);
    menu.css({top: e.pageY, left: e.pageX});
    menu.show();
  });

  history('replaceState', file);
  window.onpopstate = function(e) {
    var s = e.state;
    if (s) {
      loadAndShowSource(s.file, s.line);
    }
  }
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

function hideOnClickOff(menu) {
  $('body').click(function(e) {
    if (!$(e.target).closest(code).length) {
      menu.hide();
    }
  });
}

function selectedRange() {
  return window.getSelection().getRangeAt(0);
}

function isRangeWithinElement(range, elem) {
  return range.commonAncestorContainer.parentElement == elem[0];
}

function filterModes(menu, range) {
  menu.find('li').each(function() {
    $(this).toggle($(this).data('mode').menu(range));
  });
}

function queryAction(mode, start, end) {
  writeOutput(message.wait);
  return query(mode, pos(currentFile, start, end), 'plain')
    .done(function(data) {
      writeOutput(data);
    })
    .fail(function(e) {
      writeOutput(message.error);
    });
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

function writeOutput(text) {
  appendLinkified(out.empty(), text).trigger('change');
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
      var link = sourceLink(file, fromLine, arrow + rest, line);
      element.append(link).append('\n');
      continue;
    }
    if (match = singleAddress.exec(line)) {
      var file = match[1];
      var lineNo = match[2];
      var col = match[3];
      var rest = match[4];
      var link = sourceLink(file, lineNo, arrow + rest, line);
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

function sourceLink(file, line, text, tooltip) {
  var link = $('<a>').attr('title', tooltip).text(text);
  link.click(function() {
    loadAndShowSource(file, line);
    history('pushState', file, line);
  });
  return link;
}

function loadAndShowSource(file, line) {
  return loadFile(file)
    .done(function(src) {
      replaceSource(src);
      setCurrentFile(file);
      jumpTo(line);
    })
    .fail(function() {
      writeOutput(message.error);
    });
}

function history(method, file, line) {
  var url = 'source?' + $.param({'file': file});
  if (line) {
    url += '#L'+line;
  }
  window.history[method]({'file': file, 'line': line}, '', url);
}

function loadFile(path) {
  return $.get('file?' + $.param({'path': path}));
}

function pos(file, start, end) {
  var p = file + ':#' + start;
  if (start != end) {
    p += ',#' + end;
  }
  return p;
}

function replaceSource(src) {
  code.text(src);
  showNumbers(countLines(src));
}

function showNumbers(n) {
  nums.empty();
  for (var i = 1; i <= n; i++) {
    nums.append($('<span>').attr('id', 'L'+i).text(i)).append('<br>');
  }
}

function setCurrentFile(path) {
  currentFile = path;
  $('h1').text('Source file ' + path);
  document.title = path + ' - ' + title;
}

function jumpTo(line) {
  if (!line) {
    $('#content').scrollTop(0);
    return;
  }
  $('#L'+line)[0].scrollIntoView(true);
}

function countLines(s) {
  return (s.match(/\n/g)||[]).length;
}

function modeMenu() {
  var m = $('<ul class="menu">').hide();
  $.each(modes, function(i, mode) {
    var item = $('<li>').text(mode.name).attr('title', mode.desc)
      .data('mode', mode)
      .click(function() {
        m.trigger('select', mode.id);
      });
    m.append(item);
  });
  return m;
}

function makeQueryButton(elem, modeId) {
  var mode = $.grep(modes, function(m) {
    return m.id == modeId;
  })[0];
  elem.text(mode.name).attr('title', mode.desc)
    .addClass('button')
    .click(function() {
      queryAction(mode.id, 0, 0);
    });
}

function getByteOffsets(s, range) {
  var a = getUTF8Length(s, 0, range.startOffset);
  var b = getUTF8Length(s, range.startOffset, range.endOffset);
  return {startOffset: a, endOffset: a+b};
}

// From http://stackoverflow.com/a/12206089
function getUTF8Length(s, start, end) {
  var len = 0;
  for (var i = start; i < end; i++) {
    var code = s.charCodeAt(i);
    if (code <= 0x7f) {
      len += 1;
    } else if (code <= 0x7ff) {
      len += 2;
    } else if (code >= 0xd800 && code <= 0xdfff) {
      // Surrogate pair: These take 4 bytes in UTF-8 and 2 chars in UCS-2
      // (Assume next char is the other [valid] half and just skip it)
      len += 4;
      i++;
    } else if (code < 0xffff) {
      len += 3;
    } else {
      len += 4;
    }
  }
  return len;
}

return {
  init: init,
  makeQueryButton: makeQueryButton
};

})();
