// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

const guru = (() => {

  'use strict';

  const modes = [
    {
      id: 'definition', menu: false, name: 'Definition',
      desc: 'Show the definition of the selected identifier.'
    },
    {
      id: 'describe', menu: true, name: 'Describe',
      desc: 'Describe the selected syntax, its kind, type and methods.'
    },
    {
      id: 'callees', menu: true, name: 'Call targets',
      desc: 'Show possible callees of the function call at the current point.'
    },
    {
      id: 'callers', menu: true, name: 'Callers',
      desc: 'Show the set of callers of the function containing the ' +
          'current point.'
    },
    {
      id: 'callstack', menu: true, name: 'Call stack',
      desc: 'Show an arbitrary path from a root of the call graph to the ' +
          'function containing the current point.'
    },
    {
      id: 'freevars', menu: true, name: 'Free variables',
      desc: 'Enumerate the free variables of the current selection.'
    },
    {
      id: 'implements', menu: true, name: 'Implements',
      desc: 'Describe the \'implements\' relation for types in the package ' +
          'containing the current point.'
    },
    {
      id: 'peers', menu: true, name: 'Channel peers',
      desc: 'Enumerate the set of possible corresponding sends/receives for ' +
          'this channel receive/send operation.'
    },
    {
      id: 'pointsto', menu: true, name: 'Points to',
      desc: 'Show what the selected expression points to.'
    },
    {
      id: 'referrers', menu: true, name: 'Referrers',
      desc: 'Enumerate all references to the object denoted by the selected ' +
          'identifier.'
    },
    {
      id: 'whicherrs', menu: true, name: 'Which errors',
      desc: 'Show globals, constants and types to which the selected' +
          'expression (of type \'error\') may refer.'
    }
  ];

  const message = {
    wait: 'Consulting the guru ...',
    error: 'An error occurred.'
  };

  const title = 'Go source code guru';

  let currentFile;
  let nums, code, out;

  function init(source, output, file) {
    makeSourceView(source);
    nums = source.find('.nums');
    code = source.find('.lines');
    out = output.addClass('out');
    currentFile = file;

    loadAndShowSource(file);

    const menu = modeMenu();
    $('body').append(menu);
    hideOnEscape(menu, out);
    hideOnClickOff(menu);
    code.mouseup(e => {
      if (menu.is(':visible')) {
        menu.hide();
        return;
      }

      const range = selectedRange();
      if (!isRangeWithinElement(range, code)) {
        return;
      }

      insertSelectionMarks(range);
      const sel = selectionMarkOffsets();
      detachSelectionMarks();

      menu.unbind('select').on('select', (e, mode) => {
        menu.hide();
        const b = getByteOffsets(code.text(), sel);
        queryAction(mode.id, b.startOffset, b.endOffset);
      });
      filterModes(menu, sel);
      menu.css({top: e.pageY, left: e.pageX});
      menu.show();
    });

    history('replaceState', file);
    window.onpopstate = e => {
      const s = e.state;
      if (s) {
        loadAndShowSource(s.file, s.line, s.sel);
      }
    };
  }

  const marker = '\0';
  const startMark = makeSelectionMark('start');
  const endMark = makeSelectionMark('end');

  function makeSelectionMark(type) {
    return $('<span class="mark">').addClass(type).text(marker)[0];
  }

  function insertSelectionMarks(range) {
    const s = range.cloneRange();
    const e = range.cloneRange();
    s.collapse(true);
    e.collapse(false);
    s.insertNode(startMark);
    e.insertNode(endMark);
  }

  function detachSelectionMarks() {
    code.detach('.mark');
  }

  function selectionMarkOffsets() {
    const marked = code.text();
    return {
      startOffset: marked.indexOf(marker),
      endOffset: marked.lastIndexOf(marker) - 1
    };
  }

  const ESC = 27;

  function hideOnEscape(menu, out) {
    $(document).keyup(e => {
      if (e.keyCode === ESC) {
        if (menu.is(':visible')) {
          menu.hide();
          return;
        }
        out.trigger('esc');
      }
    });
  }

  function hideOnClickOff(menu) {
    $('body').click(e => {
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
    queryApplicableModes(range, modeIds => {
      menu.find('li').each(function () {
        const mode = $(this).data('mode');
        const applicable = $.inArray(mode.id, modeIds) >= 0;
        $(this).toggleClass('disabled', !applicable);
      });
    });
  }

  function queryApplicableModes(range, callback) {
    const b = getByteOffsets(code.text(), range);
    query('what', pos(currentFile, b.startOffset, b.endOffset), 'json')
        .done(data => {
          callback(data.what.modes);
        })
        .fail(e => {
          console.log(e);
        });
  }

  function queryAction(mode, start, end) {
    writeOutput(message.wait);
    return query(mode, pos(currentFile, start, end), 'plain')
        .done(data => {
          writeOutput(data);
        })
        .fail(() => {
          writeOutput(message.error);
        });
  }

  function query(mode, pos, format) {
    const data = {
      mode: mode,
      pos: pos,
      format: format
    };
    const get = (format === 'json') ? $.getJSON : $.get;
    return get('query', data);
  }

  function writeOutput(text) {
    appendLinkified(out.empty(), text).trigger('change');
  }

  // file:line.col-line.col:
  const rangeAddress = /(.*):([0-9]+)\.([0-9]+)-([0-9]+)\.([0-9]+): (.*)/;
  // file:line:col:
  const singleAddress = /(.*):([0-9]+):([0-9]+): (.*)/;
  // -:
  const noAddress = /-: (.*)/;

  function appendLinkified(element, text) {
    let match;
    const arrow = '▶ ';
    const lines = text.split('\n');
    const n = lines.length;
    for (let i = 0; i < n; i++) {
      const line = lines[i];
      match = rangeAddress.exec(line)
      if (match) {
        const file = match[1];
        const sel = {
          fromLine: parseInt(match[2], 10),
          fromCol: parseInt(match[3], 10),
          toLine: parseInt(match[4], 10),
          toCol: parseInt(match[5], 10)
        };
        const rest = match[6];
        const link = sourceLink(file, sel, arrow + rest, line);
        element.append(link).append('\n');
        continue;
      }
      match = singleAddress.exec(line)
      if (match) {
        const file = match[1];
        const sel = {
          fromLine: parseInt(match[2], 10),
          fromCol: parseInt(match[3], 10),
          toLine: parseInt(match[2], 10),
          toCol: parseInt(match[3], 10)
        };
        const rest = match[4];
        const link = sourceLink(file, sel, arrow + rest, line);
        element.append(link).append('\n');
        continue;
      }
      match = noAddress.exec(line)
      if (match) {
        const rest = match[1];
        element.append('  ' + rest + '\n');
        continue;
      }
      element.append(line + '\n');
    }
    return element;
  }

  function sourceLink(file, sel, text, tooltip) {
    const link = $('<a>').attr('title', tooltip).text(text);
    link.click(() => {
      loadAndShowSource(file, sel.fromLine, sel);
      history('pushState', file, sel.fromLine, sel);
    });
    return link;
  }

  function loadAndShowSource(file, line, sel) {
    return loadFile(file, sel)
        .done(src => {
          replaceSource(src);
          setCurrentFile(file);
          jumpTo(line);
        })
        .fail(() => {
          writeOutput(message.error);
        });
  }

  function history(method, file, line, sel) {
    window.history[method]({'file': file, 'line': line, 'sel': sel}, '', sourceURL(file, line));
  }

  function sourceURL(file, line) {
    const url = 'source?' + $.param({'file': file});
    if (line) {
      return url + '#L' + line;
    }
    return url;
  }

  function loadFile(path, sel) {
    const params = {'path': path};
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
    const p = file + ':#' + start;
    if (start === end) {
      return p;
    }
    return p + ',#' + end;
  }

  function replaceSource(src) {
    code.html(src);
    showNumbers(countLines(src));
  }

  function showNumbers(n) {
    nums.empty();
    for (let i = 1; i <= n; i++) {
      nums.append($('<span>').attr('id', 'L' + i).text(i)).append('<br>');
    }
  }

  function setCurrentFile(path) {
    currentFile = path;
    $('h1').text('Source file ' + path);
    document.title = path + ' - ' + title;
  }

  function jumpTo(line) {
    const content = $('#content');
    if (!line) {
      content.scrollTop(0);
      return;
    }
    const l = $('#L' + line);
    l.addClass('selection', 100).delay(300).removeClass('selection', 600);

    // scroll into view with 3 lines of padding above
    content.animate({
      scrollTop: content.scrollTop() + l.offset().top -
          l.height() * 3 - $('#top').height()
    });
  }

  function countLines(s) {
    return (s.match(/\n/g) || []).length;
  }

  function modeMenu() {
    const m = $('<ul class="menu">').hide();
    $.each(modes, (i, mode) => {
      if (!mode.menu) {
        return;
      }
      const item = $('<li>').text(mode.name).attr('title', mode.desc)
          .data('mode', mode)
          .click(function () {
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

  function getByteOffsets(s, range) {
    const a = getUTF8Length(s, 0, range.startOffset);
    const b = getUTF8Length(s, range.startOffset, range.endOffset);
    return {startOffset: a, endOffset: a + b};
  }

  // From http://stackoverflow.com/a/12206089
  function getUTF8Length(s, start, end) {
    let len = 0;
    for (let i = start; i < end; i++) {
      const code = s.charCodeAt(i);
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
    init: init
  };

})();
