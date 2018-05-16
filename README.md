# random pick my favorite article from getpocket.com

launch server

    export CONSUMER_KEY={your-get-pocket-consumer-key}
    export SECRET_KEY={random}
    export ROOT_URL={your-root-url}

    gunicorn app:app --bind 127.0.0.1:8000 --workers 4 --access-logfile -

and open ROOT_URL with your browser.

http://pick.woosum.net

# 왜?

기사를 보다 나중에 보거나, 시간이 흐른 뒤에도 볼만할 글들은 pocket에서 즐겨찾기 항목으로 저장하는데,
이제 양이 많아져서 심심할 때 그 글중에서 아무거나 읽는 기능을 그냥 만들어 봄..

https://getpocket.com/random 이 있기는 하지만, 이건 내가 원하는 기능이 아니여서 간단히..
