"""MagicSession"""

import json
import requests


class MagicSession(object):
    """MagicSession"""

    def __init__(self, base_url, namespace):
        self.current_session = requests.Session()
        self.base_url = base_url
        self.namespace = namespace
        self.session_token = None

    def new_session(self):
        """fork new session"""
        return MagicSession(self.base_url, self.namespace)

    def bind_token(self, token):
        self.session_token = token

    def header(self):
        header = {
            "X-Namespace": self.namespace
        }

        if self.session_token:
            header["Authorization"] = 'Bearer %s' % self.session_token

        return header

    def post(self, url, params):
        """Post"""
        ret = None
        try:
            response = self.current_session.post('%s%s' % (self.base_url, url), headers=self.header(), json=params)
            ret = json.loads(response.text)
        except ValueError as except_value:
            print(except_value)

        return ret

    def get(self, url, params=None):
        """Get"""
        ret = None
        try:
            response = self.current_session.get('%s%s' % (self.base_url, url), headers=self.header(), params=params)
            ret = json.loads(response.text)
        except ValueError as except_value:
            print(except_value)

        return ret

    def put(self, url, params):
        """Put"""
        ret = None
        try:
            response = self.current_session.put('%s%s' % (self.base_url, url), headers=self.header(), json=params)
            ret = json.loads(response.text)
        except ValueError as except_value:
            print(except_value)

        return ret

    def delete(self, url):
        """Delete"""
        ret = None
        try:
            response = self.current_session.delete('%s%s' % (self.base_url, url), headers=self.header())
            ret = json.loads(response.text)
        except ValueError as except_value:
            print(except_value)

        return ret

    def upload(self, url, files, params=None):
        """Upload"""
        ret = None
        try:
            response = self.current_session.post('%s%s' % (self.base_url, url),
                                                 headers=self.header(), params=params, files=files)
            ret = json.loads(response.text)
        except ValueError as except_value:
            print(except_value)

        return ret

    def download(self, url, dst_file, params=None):
        """Download"""
        ret = None
        try:
            response = self.current_session.get('%s%s' % (self.base_url, url), headers=self.header(), params=params)
            with open(dst_file, "wb") as code:
                code.write(response.content)

            ret = dst_file
        except ValueError as except_value:
            print(except_value)

        return ret

