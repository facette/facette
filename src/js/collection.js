
/* Collection */

function collectionDelete(id) {
    return $.ajax({
        url: urlPrefix + '/library/collections/' + id,
        type: 'DELETE'
    });
}

function collectionList(query) {
    return $.ajax({
        url: urlPrefix + '/library/collections',
        type: 'GET',
        data: query,
        dataType: 'json'
    });
}

function collectionLoad(id) {
    return $.ajax({
        url: urlPrefix + '/library/collections/' + id,
        type: 'GET',
        dataType: 'json'
    });
}

function collectionSave(id, query, mode) {
    var url = '/library/collections',
        method = 'POST';

    if (mode === SAVE_MODE_CLONE) {
        url += '?inherit=' + id;
    } else if (id !== null) {
        url += '/' + id;
        method = 'PUT';
    }

    return $.ajax({
        url: urlPrefix + url,
        type: method,
        contentType: 'application/json',
        data: JSON.stringify(query)
    });
}
