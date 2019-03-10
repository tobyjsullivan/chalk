import axios from 'axios';

import {SessionState} from '../domain';
import { List } from 'immutable';

const ENDPOINT_CREATE_SESSION = '/sessions';

type CreateSessionResponse = SessionResponse;
type GetSessionResponse = SessionResponse | ErrorResponse;

interface SessionResponse {
  error: undefined,
  session: {
    id: string,
    pages: ReadonlyArray<string>,
  },
}

interface ErrorResponse {
  error: string,
  session: undefined,
}

export async function createSession(apiUrl: string): Promise<SessionState> {
  const {data}: {data: CreateSessionResponse} = await axios.post(apiUrl + ENDPOINT_CREATE_SESSION);

  return {
    id: data.session.id,
    pages: List(data.session.pages).map(pageId => ({id: pageId})),
  };
}

export async function getSession(apiUrl: string, sessionId: string): Promise<SessionState | null> {
  const safeSessionId = encodeURIComponent(sessionId);
  const {data}: {data: GetSessionResponse} = await axios.get(apiUrl + ENDPOINT_CREATE_SESSION + '/' + safeSessionId);
  if (data.error !== undefined) {
    return null;
  }

  return {
    id: data.session.id,
    pages: List(data.session.pages).map(pageId => ({id: pageId})),
  };
}
