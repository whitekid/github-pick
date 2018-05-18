import logging
import os

from flask import Flask, redirect, request, session

import pocket

app = Flask(__name__)
app.secret_key = os.environ.get('SECRET_KEY', 'secret-key')

pocket.CONSUMER_KEY = os.environ['CONSUMER_KEY']
ROOT_URL = os.environ.get('ROOT_URL', 'http://pick.woosum.net')

LOG = logging.getLogger(__name__)


@app.route('/')
def index():
    request_token = session.get('REQUEST_TOKEN')
    access_token = session.get('ACCESS_TOKEN')
    unread = request.args.get('unread')

    if not request_token:
        request_token, authorized_url = pocket.get_authorized_url(
            f'{ROOT_URL}/auth?unread={unread}')
        session['REQUEST_TOKEN'] = request_token
        return redirect(authorized_url)

    if request_token:
        if access_token:
            return get_random_pick(access_token, unread)
        else:
            return redirect(f'{ROOT_URL}/auth?unread={unread}')

    return ''


@app.route('/auth')
def auth():
    request_token = session.get('REQUEST_TOKEN')
    if not request_token:
        LOG.info('No request token')
        return redirect(ROOT_URL)

    access_token = session.get('ACCESS_TOKEN')
    if not access_token:
        access_token, _ = pocket.get_access_token(request_token)
        if not access_token:
            del session['REQUEST_TOKEN']
            return redirect(ROOT_URL)
        session['ACCESS_TOKEN'] = access_token

    return get_random_pick(access_token, request.args.get('unread'))


def get_random_pick(access_token: str, unread=None):
    state = 'all'
    favorite = True
    if unread:
        state = 'unread'
        favorite = False

    url = pocket.random_pick(access_token, state, favorite)
    if url:
        return redirect(url)
    else:
        del session['ACCESS_TOKEN']
        del session['REQUEST_TOKEN']
        return redirect(ROOT_URL)


if __name__ == '__main__':
    logging.getLogger().setLevel(logging.DEBUG)
    app.run(host='0.0.0.0', port=5000, debug=True)
