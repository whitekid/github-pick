import logging
import random

import requests

CONSUMER_KEY = None

LOG = logging.getLogger(__name__)


def get_authorized_url(redirect_uri: str):
    # get request token
    resp = requests.post(
        'https://getpocket.com/v3/oauth/request',
        headers={'X-Accept': 'application/json'},
        json={
            'consumer_key': CONSUMER_KEY,
            'redirect_uri': redirect_uri,
        })
    resp.raise_for_status()

    request_token = resp.json()['code']
    return request_token, \
        f'https://getpocket.com/auth/authorize?request_token={request_token}&' \
        f'redirect_uri={redirect_uri}'


def get_access_token(request_token: str):
    resp = requests.post(
        'https://getpocket.com/v3/oauth/authorize',
        headers={'X-Accept': 'application/json'},
        json={
            'consumer_key': CONSUMER_KEY,
            'code': request_token,
        })

    if resp.status_code in (400, 403):
        return None, None

    return resp.json()['access_token'], resp.json()['username']


def get(access_token: str,
        state: str = 'unread',
        sort: str = 'newest',
        detail_type: str = 'simple',
        favorite: bool = None):
    param = {
        'consumer_key': CONSUMER_KEY,
        'access_token': access_token,
        'state': state,
        'sort': sort,
        'detailType': detail_type,
    }

    if favorite is not None:
        param['favorite'] = 1 if favorite else 0

    resp = requests.post(
        'https://getpocket.com/v3/get',
        headers={'X-Accept': 'application/json'},
        json=param)
    resp.raise_for_status()

    return resp.json()


def get_random_favorite(access_token: str):
    try:
        items = get(access_token, state='all', favorite=True)['list']
    except requests.exceptions.HTTPError as err:
        if err.status_code == 400:
            return None
        raise
    item_id = random.choice(list(items.keys()))

    return f'https://getpocket.com/a/read/{item_id}'
