import * as React from 'react'
import fetch, { Headers, Request, Response } from 'node-fetch';

global.React = React

if (!globalThis.fetch) {
  globalThis.fetch = fetch;
  globalThis.Headers = Headers;
  globalThis.Request = Request;
  globalThis.Response = Response;
}
