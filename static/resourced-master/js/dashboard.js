var ResourcedMaster = ResourcedMaster || {};

ResourcedMaster.globals = {};
ResourcedMaster.globals.currentCluster = {};
ResourcedMaster.globals.TSEventLines = [];
ResourcedMaster.globals.TSEventBands = [];
ResourcedMaster.globals.TSEventLineColors = [];
ResourcedMaster.globals.TSEventBandColors = [];
ResourcedMaster.globals.TSEventBandTextColors = [];

ResourcedMaster.url = {};
ResourcedMaster.url.getParams = function(sParam) {
    var sPageURL = decodeURIComponent(window.location.search.substring(1)),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : sParameterName[1];
        }
    }
};

ResourcedMaster.users = {};
ResourcedMaster.users.logout = function() {
    $.removeCookie('resourcedmaster-session', { path: '/' });
    window.location = '/login';
};

ResourcedMaster.hosts = {};
ResourcedMaster.hosts.get = function(accessToken, options) {
    var path = '/api/hosts';
    var getParams = '';

    if('query' in options) {
        getParams = getParams + 'q=' + options.query;
    }
    if('count' in options) {
        getParams = getParams + '&count=' + options.count;
    }

    return $.ajax({
        url: path + '?' + getParams,
        beforeSend: function(xhr) {
            xhr.setRequestHeader('Authorization', 'Basic ' + window.btoa(accessToken + ':'));
        },
        success: options.successCallback || null
    });
};

ResourcedMaster.logs = {};
ResourcedMaster.logs.get = function(accessToken, options) {
    var path = options.path || '/api/logs';
    var getParams = '';

    if('query' in options && options.query) {
        getParams = getParams + 'q=' + options.query;
    }
    if('from' in options) {
        getParams = getParams + '&from=' + options.from;
    }
    if('to' in options) {
        getParams = getParams + '&to=' + options.to;
    }

    return $.ajax({
        url: path + '?' + getParams,
        beforeSend: function(xhr) {
            xhr.setRequestHeader('Authorization', 'Basic ' + window.btoa(accessToken + ':'));
        },
        success: options.successCallback || null
    });
};
ResourcedMaster.logs.render = function(olderOrNewer, logsJSON, itemsPerPage, ulElem, prevElem, nextElem) {
    var shouldPaginate = logsJSON.length >= itemsPerPage;

    var storedOlderOrNewer = ulElem.data('older-or-newer');

    // Newer means we are paginating smaller array index number.
    if(olderOrNewer == 'newer') {
        if (itemsPerPage >= 0) {
            itemsPerPage = itemsPerPage * -1;
        }
    }

    var index = ulElem.data('index') % logsJSON.length || 0;

    var newIndex = index + itemsPerPage;
    if(newIndex < 0) {
        newIndex = 0;
    }

    // If user is switching back and forth between older and newer arrows,
    // paginate one step further.
    if(storedOlderOrNewer != olderOrNewer) {
        index = index + itemsPerPage;
        newIndex = index + itemsPerPage;
    }

    ulElem.data('index', newIndex);
    ulElem.data('older-or-newer', olderOrNewer);

    var logsJSONForDisplay = [];

    if(shouldPaginate) {
        logsJSONForDisplay = logsJSON.slice(Math.min(index, newIndex), Math.max(index, newIndex));
    } else {
        logsJSONForDisplay = logsJSON;

        prevElem.prop('disabled', true);
        nextElem.prop('disabled', true);
    }

    ulElem.html($.map(logsJSONForDisplay, function(val) {
        var tags = '';

        for (var prop in val.Tags) {
            // skip loop if the property is from prototype
            if(!val.Tags.hasOwnProperty(prop)) continue;

            var tag = '<a data-clause="tags.' + prop + '=\'' + val.Tags[prop] + '\'">' + prop + ": " + val.Tags[prop] + '</a>';
            tags = tags + tag;
        }

        return '<li>' +
            '<div class="logline">' + val.Logline + '</div>' +
            '<div class="hostname"><a data-clause="hostname=\'' + val.Hostname + '\'">' + val.Hostname + '</a></div>' +
            '<div class="tags">' + tags + '</div>' +
        '</li>';
    }).join(''));
};

ResourcedMaster.graphs = {};
ResourcedMaster.graphs.ajax = function(accessToken, options) {
    var path = '/api/graphs';
    var getParams = '';
    var method = 'GET';
    var dataJSON = '';

    if('method' in options) {
        method = options.method;
    }
    if('id' in options) {
        path = path + '/' + options.id;
    }
    if('metrics' in options) {
        path = path + '/metrics';
    }
    if('data' in options) {
        dataJSON = JSON.stringify(options.data);
    }

    return $.ajax({
        url: path + '?' + getParams,
        contentType : 'application/json',
        beforeSend: function(xhr) {
            xhr.setRequestHeader('Authorization', 'Basic ' + window.btoa(accessToken + ':'));
        },
        method: method,
        data: dataJSON,
        success: options.successCallback || null
    });
};

ResourcedMaster.daterange = {};
ResourcedMaster.daterange.defaultSettings = {
    'timePicker': true,
    'timePickerSeconds': true,
    'autoApply': true,
    'ranges': {
        '5 minutes': [moment().subtract(5, 'minutes'), moment()],
        '10 minutes': [moment().subtract(10, 'minutes'), moment()],
        '15 minutes': [moment().subtract(15, 'minutes'), moment()],
        '30 minutes': [moment().subtract(30, 'minutes'), moment()],
        '60 minutes': [moment().subtract(60, 'minutes'), moment()],
        '2 hours': [moment().subtract(2, 'hours'), moment()],
        '3 hours': [moment().subtract(3, 'hours'), moment()],
        '6 hours': [moment().subtract(6, 'hours'), moment()],
        '12 hours': [moment().subtract(12, 'hours'), moment()],
        '24 hours': [moment().subtract(24, 'hours'), moment()],
        '2 days': [moment().subtract(2, 'days'), moment()],
        '3 days': [moment().subtract(3, 'days'), moment()],
        '7 days': [moment().subtract(7, 'days'), moment()]
    },
    'defaultDate': moment(),
    'startDate': moment().subtract(15, 'minutes'),
    'endDate': moment(),
    'locale': {
        'format': 'YYYY/MM/DD hh:mm:ss A'
    },
    'opens': 'left'
};


ResourcedMaster.metrics = {};
ResourcedMaster.metrics.get = function(accessToken, metricID, options) {
    var path = '/api/metrics/' + metricID;
    var getParams = '';

    if('host' in options) {
        path = path + '/hosts/' + options.host;
    }
    if('shortAggrInterval' in options) {
        path = path + '/' + options.shortAggrInterval;
    }
    if('from' in options) {
        getParams = getParams + 'From=' + options.from;
    }
    if('to' in options) {
        getParams = getParams + '&To=' + options.to;
    }
    if('aggr' in options) {
        getParams = getParams + '&aggr=' + options.aggr;
    }

    return $.ajax({
        url: path + '?' + getParams,
        beforeSend: function(xhr) {
            xhr.setRequestHeader("Authorization", "Basic " + window.btoa(accessToken + ':'));
        },
        success: options.successCallback || null,
        error: options.errorCallback || function() {
            if(toastr) {
                toastr.error('API for Metric(ID: ' + metricID + ') failed to return data');
            }
        }
    });
};
ResourcedMaster.metrics.renderOneChart = function(accessToken, metricID, eventLines, eventLineColors, eventBands, eventBandColors, eventBandTextColors, options) {
    options.successCallback = function(result) {
        if((!result || !result.data || result.data.length == 0) && toastr) {
            toastr.warning('API for Metric(ID: ' + metricID + ') returned no data');
        }

        if(result.constructor != Array) {
            result = [result];
        }

        // Check if result is aggregated data.
        // If so, then the result payload need to be rearranged.
        if(result[0] && result[0]['data'][0]) {
            var firstValue = result[0]['data'][0][1];

            if(typeof firstValue === 'object') {
                var avgResult = {'name': result[0]['name'] + ' avg', 'data': []};
                var maxResult = {'name': result[0]['name'] + ' max', 'data': []};
                var minResult = {'name': result[0]['name'] + ' min', 'data': []};
                var sumResult = {'name': result[0]['name'] + ' sum', 'data': []};

                for(i = 0; i < result.length; i++) {
                    var data = result[i]['data'];

                    for(j = 0; j < data.length; j++) {
                        var eachData = result[i]['data'][j];

                        avgResult['data'].push([eachData[0], eachData[1]['avg']]);
                        maxResult['data'].push([eachData[0], eachData[1]['max']]);
                        minResult['data'].push([eachData[0], eachData[1]['min']]);
                        sumResult['data'].push([eachData[0], eachData[1]['sum']]);
                    }
                }

                result = [avgResult, maxResult, minResult, sumResult];
            }
        }

        var hcPlotLines = [];
        for(i = 0; i < eventLines.length; i++) {
            var plotEventSettings = {
                color: eventLineColors[i],
                width: 1,
                value: eventLines[i].CreatedFrom,
                id: eventLines[i].ID,
                dashStyle: 'longdashdot',
                label: {
                    text: eventLines[i].Description,
                    style: {
                        color: eventLineColors[i]
                    }
                }
            };
            hcPlotLines[i] = plotEventSettings;
        }

        var hcPlotBands = [];
        for(i = 0; i < eventBands.length; i++) {
            var plotEventSettings = {
                color: eventBandColors[i],
                from: eventBands[i].CreatedFrom,
                to: eventBands[i].CreatedTo,
                id: eventBands[i].ID,
                dashStyle: 'longdashdot',
                label: {
                    text: eventBands[i].Description,
                    style: {
                        color: eventBandTextColors[i],
                        fontWeight: 'bold'
                    }
                }
            };
            hcPlotBands[i] = plotEventSettings;
        }

        var hcOptions = {
            chart: {
                width: options.containerDOM.width(),
                height: options.height || ResourcedMaster.highcharts.defaultHeight
            },
            title: {
                text: options.title || ''
            },
            series: result,
            xAxis: {
                plotLines: hcPlotLines,
                plotBands: hcPlotBands
            }
        };
        if(options.onLoad) {
            if(!hcOptions.chart.events) {
                hcOptions.chart.events = {};
            }
            hcOptions.chart.events.load = options.onLoad;
        }

        options.containerDOM.highcharts(hcOptions);
    };

    return ResourcedMaster.metrics.get(accessToken, metricID, options);
};
ResourcedMaster.metrics.getEvents = function(accessToken, eventType, options) {
    var path = '/api/events/' + eventType;
    var getParams = '';

    if('from' in options) {
        getParams = getParams + 'From=' + options.from;
    }
    if('to' in options) {
        getParams = getParams + '&To=' + options.to;
    }

    return $.ajax({
        url: path + '?' + getParams,
        beforeSend: function(xhr) {
            xhr.setRequestHeader("Authorization", "Basic " + window.btoa(accessToken + ':'));
        },
        success: options.successCallback || null
    });
};

ResourcedMaster.metrics.getEventsLastXRange = function(count, unit, doneCallback) {
    $.when(
        ResourcedMaster.metrics.getEvents(ResourcedMaster.globals.AccessToken, 'line', {
            'from': moment().subtract(count, unit),
            'to': moment(),
            'successCallback': function(result) {
                ResourcedMaster.globals.TSEventLines = result;
            }
        }),
        ResourcedMaster.metrics.getEvents(ResourcedMaster.globals.AccessToken, 'band', {
            'from': moment().subtract(count, unit),
            'to': moment(),
            'successCallback': function(result) {
                ResourcedMaster.globals.TSEventLines = result;
            }
        })
    ).done(function(a1, a2) {
        // a1 and a2 are arguments resolved for the page1 and page2 ajax requests, respectively.
        // Each argument is an array with the following structure: [ data, statusText, jqXHR ]
        ResourcedMaster.globals.TSEventLines = a1[0];
        ResourcedMaster.globals.TSEventBands = a2[0];

        ResourcedMaster.globals.TSEventLineColors = randomColor({hue: 'green', luminosity: 'light', count: ResourcedMaster.globals.TSEventLines.length});
        ResourcedMaster.globals.TSEventBandColors = randomColor({hue: 'yellow', luminosity: 'light', count: ResourcedMaster.globals.TSEventBands.length});
        ResourcedMaster.globals.TSEventBandTextColors = randomColor({hue: 'green', luminosity: 'dark', count: ResourcedMaster.globals.TSEventBands.length});

        if(doneCallback) {
            doneCallback(a1, a2);
        }
    });
};

ResourcedMaster.metrics.get1dayEvents = function(doneCallback) {
    return ResourcedMaster.metrics.getEventsLastXRange(1, 'days', doneCallback);
};


ResourcedMaster.highcharts = {};
ResourcedMaster.highcharts.defaultHeight = 300;

// ---------------------------------------
// Highchart Settings
// ---------------------------------------
// Highcharts.createElement('link', {
//     href: '//fonts.googleapis.com/css?family=Unica+One',
//     rel: 'stylesheet',
//     type: 'text/css'
// }, null, document.getElementsByTagName('head')[0]);

Highcharts.setOptions({
    global: {
        useUTC: false
    }
});

Highcharts.theme = {
    colors: ["#2b908f", "#90ee7e", "#f45b5b", "#7798BF", "#aaeeee", "#ff0066", "#eeaaee", "#55BF3B", "#DF5353", "#7798BF", "#aaeeee"],
    chart: {
        backgroundColor: {
            linearGradient: { x1: 0, y1: 0, x2: 1, y2: 1 },
            stops: [
                [0, '#2a2a2b'],
                [1, '#3e3e40']
            ]
        },
        style: {
            fontFamily: "'Lato', sans-serif"
        },
        plotBorderColor: '#606063',
        type: 'spline',
        animation: Highcharts.svg, // don't animate in old IE
        height: ResourcedMaster.highcharts.defaultHeight
    },
    title: {
        style: {
            color: '#E0E0E3',
            fontSize: '20px'
        }
    },
    subtitle: {
        style: {
            color: '#E0E0E3'
        }
    },
    xAxis: {
        gridLineColor: '#707073',
        labels: {
            style: {
                color: '#E0E0E3'
            }
        },
        lineColor: '#707073',
        minorGridLineColor: '#505053',
        tickColor: '#707073',
        title: {
            style: {
                color: '#A0A0A3'
            }
        },
        type: 'datetime',
        dateTimeLabelFormats: { // don't display the dummy year
            month: '%e. %b',
            year: '%b'
        }
    },
    yAxis: {
        gridLineColor: '#707073',
        labels: {
            style: {
                color: '#E0E0E3'
            }
        },
        lineColor: '#707073',
        minorGridLineColor: '#505053',
        tickColor: '#707073',
        tickWidth: 1,
        title: {
            style: {
                color: '#A0A0A3'
            },
            text: ''
        }
    },
    tooltip: {
        backgroundColor: 'rgba(0, 0, 0, 0.85)',
        style: {
            color: '#F0F0F0'
        },
        shared: true,
        crosshairs: true
    },
    exporting: {
        enabled: false
    },
    plotOptions: {
        series: {
            dataLabels: {
                color: '#B0B0B3'
            },
            marker: {
                lineColor: '#333'
            }
        },
        boxplot: {
            fillColor: '#505053'
        },
        candlestick: {
            lineColor: 'white'
        },
        errorbar: {
            color: 'white'
        },
        spline: {
            marker: {
                enabled: true
            }
        }
    },
    legend: {
        itemStyle: {
            color: '#E0E0E3'
        },
        itemHoverStyle: {
            color: '#FFF'
        },
        itemHiddenStyle: {
            color: '#606063'
        }
    },
    credits: {
        style: {
            color: '#666'
        }
    },
    labels: {
        style: {
            color: '#707073'
        }
    },

    drilldown: {
        activeAxisLabelStyle: {
            color: '#F0F0F3'
        },
        activeDataLabelStyle: {
            color: '#F0F0F3'
        }
    },

    navigation: {
        buttonOptions: {
            symbolStroke: '#DDDDDD',
            theme: {
                fill: '#505053'
            }
        }
    },

    // scroll charts
    rangeSelector: {
        buttonTheme: {
            fill: '#505053',
            stroke: '#000000',
            style: {
                color: '#CCC'
            },
            states: {
                hover: {
                    fill: '#707073',
                    stroke: '#000000',
                    style: {
                        color: 'white'
                    }
                },
                select: {
                    fill: '#000003',
                    stroke: '#000000',
                    style: {
                        color: 'white'
                    }
                }
            }
        },
        inputBoxBorderColor: '#505053',
        inputStyle: {
            backgroundColor: '#333',
            color: 'silver'
        },
        labelStyle: {
            color: 'silver'
        }
    },

   navigator: {
        handles: {
            backgroundColor: '#666',
            borderColor: '#AAA'
        },
        outlineColor: '#CCC',
        maskFill: 'rgba(255,255,255,0.1)',
        series: {
            color: '#7798BF',
            lineColor: '#A6C7ED'
        },
        xAxis: {
            gridLineColor: '#505053'
        }
    },

    scrollbar: {
        barBackgroundColor: '#808083',
        barBorderColor: '#808083',
        buttonArrowColor: '#CCC',
        buttonBackgroundColor: '#606063',
        buttonBorderColor: '#606063',
        rifleColor: '#FFF',
        trackBackgroundColor: '#404043',
        trackBorderColor: '#404043'
    },

    // special colors for some of the
    legendBackgroundColor: 'rgba(0, 0, 0, 0.5)',
    background2: '#505053',
    dataLabelsColor: '#B0B0B3',
    textColor: '#C0C0C0',
    contrastTextColor: '#F0F0F3',
    maskColor: 'rgba(255,255,255,0.3)'
};

Highcharts.setOptions(Highcharts.theme);

// ---------------------------------------
// Toaster Settings
// ---------------------------------------

toastr.options = {
    "closeButton": false,
    "debug": false,
    "newestOnTop": true,
    "progressBar": false,
    "positionClass": "toast-top-right",
    "preventDuplicates": true,
    "onclick": null,
    "showDuration": "300",
    "hideDuration": "1000",
    "timeOut": "5000",
    "extendedTimeOut": "1000",
    "showEasing": "swing",
    "hideEasing": "linear",
    "showMethod": "fadeIn",
    "hideMethod": "fadeOut"
}
