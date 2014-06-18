
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

function itemSave(id, itemType, query, clone) {
    var url = '/api/v1/library/' + itemType + '/',
        method = 'POST';

    clone = typeof clone == 'boolean' ? clone : false;

    if (clone) {
        url += '?inherit=' + id;
    } else if (id !== null) {
        url += id;
        method = 'PUT';
    }

    return $.ajax({
        url: urlPrefix + url,
        type: method,
        contentType: 'application/json',
        data: JSON.stringify(query)
    });
}
