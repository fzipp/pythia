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
   desc: 'Show the set of callers of the function containing the ' +
         'current point.'},
  {id: 'callgraph', menu: Filter.NO, name: 'Call graph',
   desc: 'Show the callgraph of the current program.'},
  {id: 'callstack', menu: Filter.YES, name: 'Call stack',
   desc: 'Show an arbitrary path from a root of the call graph to the ' +
         'function containing the current point.'},
  {id: 'freevars', menu: Filter.RANGE, name: 'Free variables',
   desc: 'Enumerate the free variables of the current selection.'},
  {id: 'implements', menu: Filter.YES, name: 'Implements',
   desc: 'Implements displays the "implements" relation as it pertains ' +
         'to the selected type.'},
  {id: 'peers', menu: Filter.YES, name: 'Channel peers',
   desc: 'Enumerate the set of possible corresponding sends/receives for ' +
         'this channel receive/send operation.'},
  {id: 'referrers', menu: Filter.YES, name: 'Referrers',
   desc: 'Enumerate all references to the object denoted by the selected ' +
         'identifier.'}
];

var message = {
  wait: 'Consulting the oracle ...',
  error: 'An error occurred.'
};

var title = 'Go source code oracle';

var currentFile;
var nums, code, out;

function init(source, output, file) {
  makeSourceView(source);
  nums = source.find('.nums');
  code = source.find('.lines');
  out = output.addClass('out');
  currentFile = file;

  loadAndShowSource(file);

  var menu = modeMenu();
  $('body').append(menu);
  hideOnEscape(menu, out);
  hideOnClickOff(menu);
  code.mouseup(function(e) {
    if (menu.is(':visible')) {
      menu.hide();
      return;
    }

    var range = selectedRange();
    if (!isRangeWithinElement(range, code)) {
      return;
    }

    insertSelectionMarks(range);
    var sel = selectionMarkOffsets();
    detachSelectionMarks();

    menu.unbind('select').on('select', function(e, mode) {
      menu.hide();
      var b = getByteOffsets(code.text(), sel);
      queryAction(mode.id, b.startOffset, b.endOffset);
    });
    filterModes(menu, sel);
    menu.css({top: e.pageY, left: e.pageX});
    menu.show();
  });

  history('replaceState', file);
  window.onpopstate = function(e) {
    var s = e.state;
    if (s) {
      loadAndShowSource(s.file, s.line, s.sel);
    }
  };
}

var marker = '\0';
var startMark = makeSelectionMark('start');
var endMark = makeSelectionMark('end');

function makeSelectionMark(type) {
  return $('<span class="mark">').addClass(type).text(marker)[0];
}

function insertSelectionMarks(range) {
  var s = range.cloneRange();
  var e = range.cloneRange();
  s.collapse(true);
  e.collapse(false);
  s.insertNode(startMark);
  e.insertNode(endMark);
}

function detachSelectionMarks() {
  code.detach('.mark');
}

function selectionMarkOffsets() {
  var marked = code.text();
  return {
    startOffset: marked.indexOf(marker),
    endOffset: marked.lastIndexOf(marker) - 1
  };
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
  return elem.has(range.startContainer).length &&
    elem.has(range.endContainer).length;
}

function filterModes(menu, range) {
  menu.find('li').each(function() {
    $(this).toggleClass('disabled', !$(this).data('mode').menu(range));
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
      var sel = {
        fromLine: parseInt(match[2], 10),
        fromCol: parseInt(match[3], 10),
        toLine: parseInt(match[4], 10),
        toCol: parseInt(match[5], 10)
      };
      var rest = match[6];
      var link = sourceLink(file, sel, arrow + rest, line);
      element.append(link).append('\n');
      continue;
    }
    if (match = singleAddress.exec(line)) {
      var file = match[1];
      var sel = {
        fromLine: parseInt(match[2], 10),
        fromCol: parseInt(match[3], 10),
        toLine: parseInt(match[2], 10),
        toCol: parseInt(match[3], 10)
      };
      var rest = match[4];
      var link = sourceLink(file, sel, arrow + rest, line);
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

function sourceLink(file, sel, text, tooltip) {
  var link = $('<a>').attr('title', tooltip).text(text);
  link.click(function() {
    loadAndShowSource(file, sel.fromLine, sel);
    history('pushState', file, sel.fromLine, sel);
  });
  return link;
}

function loadAndShowSource(file, line, sel) {
  return loadFile(file, sel)
    .done(function(src) {
      replaceSource(src);
      setCurrentFile(file);
      jumpTo(line);
    })
    .fail(function() {
      writeOutput(message.error);
    });
}

function history(method, file, line, sel) {
  var url = 'source?' + $.param({'file': file});
  if (line) {
    url += '#L' + line;
  }
  window.history[method]({'file': file, 'line': line, 'sel': sel}, '', url);
}

function loadFile(path, sel) {
  var params = {'path': path};
  if (sel) {
    $.extend(params, {'s': selectionParam(sel)});
  }
  return $.get('file?' + $.param(params));
}

function selectionParam(sel) {
  // line.col-line.col
  return sel.fromLine + '.' + sel.fromCol + '-' + sel.toLine + '.' + sel.toCol;
}

function pos(file, start, end) {
  var p = file + ':#' + start;
  if (start != end) {
    p += ',#' + end;
  }
  return p;
}

function replaceSource(src) {
  code.html(src);
  showNumbers(countLines(src));
}

function showNumbers(n) {
  nums.empty();
  for (var i = 1; i <= n; i++) {
    nums.append($('<span>').attr('id', 'L' + i).text(i)).append('<br>');
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
  var l = $('#L' + line);
  l.addClass('selection', 100).delay(300).removeClass('selection', 600);

  // scroll into view with 3 lines of padding above
  $('#content').animate({
    scrollTop: $('#content').scrollTop() + l.offset().top -
               l.height() * 3 - $('#top').height()
  });
}

function countLines(s) {
  return (s.match(/\n/g) || []).length;
}

function modeMenu() {
  var m = $('<ul class="menu">').hide();
  $.each(modes, function(i, mode) {
    if (mode.menu == Filter.NO) {
      return;
    }
    var item = $('<li>').text(mode.name).attr('title', mode.desc)
      .data('mode', mode)
      .click(function() {
        if (!$(this).hasClass('disabled')) {
          m.trigger('select', mode);
        }
      });
    m.append(item);
  });
  return m;
}

function makeSourceView(elem) {
  return elem.append($('<table class="code">')
    .append($('<tr>')
      .append($('<td class="nums">'))
      .append($('<td class="lines">'))));
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
  return {startOffset: a, endOffset: a + b};
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
