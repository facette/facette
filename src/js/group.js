
/* Group */

function groupDelete(id, groupType) {
    return $.ajax({
        url: '/library/' + groupType + '/' + id,
        type: 'DELETE'
    });
}

function groupList(query, groupType) {
    return $.ajax({
        url: '/library/' + groupType,
        type: 'GET',
        data: query,
        dataType: 'json'
    });
}

function groupLoad(id, groupType) {
    return $.ajax({
        url: '/library/' + groupType + '/' + id,
        type: 'GET',
        dataType: 'json'
    });
}

function groupSave(id, query, mode, groupType) {
    var url = '/library/' + groupType,
        method = 'POST';

    if (mode === SAVE_MODE_CLONE) {
        url += '?inherit=' + id;
    } else if (id !== null) {
        url += '/' + id;
        method = 'PUT';
    }

    return $.ajax({
        url: url,
        type: method,
        contentType: 'application/json',
        data: JSON.stringify(query)
    });
}
