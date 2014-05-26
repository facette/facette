
/* Item */

function itemDelete(id, itemType) {
    return $.ajax({
        url: urlPrefix + '/api/v1/library/' + itemType + '/' + id,
        type: 'DELETE'
    });
}

function itemList(query, itemType) {
    return $.ajax({
        url: urlPrefix + '/api/v1/library/' + itemType + '/',
        type: 'GET',
        data: query,
        dataType: 'json'
    });
}

function itemLoad(id, itemType) {
    return $.ajax({
        url: urlPrefix + '/api/v1/library/' + itemType + '/' + id,
        type: 'GET',
        dataType: 'json'
    });
}

function itemSave(id, itemType, query, mode) {
    var url = '/api/v1/library/' + itemType + '/',
        method = 'POST';

    if (mode === SAVE_MODE_CLONE) {
        url += '?inherit=' + id;
    } else if (mode === SAVE_MODE_VOLATILE) {
        url += '?volatile=1';
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
