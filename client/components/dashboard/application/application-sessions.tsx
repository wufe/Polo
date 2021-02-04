import { APIRequestResult } from '@/api/common';
import { ISession, SessionStatus } from '@/state/models';
import { observer } from 'mobx-react-lite';
import React from 'react';

type TProps = {
    sessions: ISession[];
}

export const ApplicationSessions = observer((props: TProps) => {

    const attachToSession = async (session: ISession) => {
        const track = await session.track();
        if (track.result === APIRequestResult.SUCCEEDED) {
            location.href = '/';
        }
    }
    const killSession = async (session: ISession) => {
        if (confirm(`You are going to delete the session. Are you sure?`)) {
            await session.kill();
        }
    }

    return <>
        <h4 className="mt-2 mb-1 text-sm text-gray-500 uppercase">Sessions:</h4>
        {props.sessions.map((session, key) =>
            <div
                key={key}
                className="items-center grid grid-cols-12 gap-2">
                <span className="text-xs uppercase col-span-1" style={{ color: colorByStatus(session.status) }}>{session.status}</span>
                <span className="text-sm mr-3 flex-grow col-span-7">{session.checkout}</span>
                <span className="text-xs uppercase text-gray-500 col-span-2">
                    Expires in {session.maxAge}s
                </span>
                <span className="col-span-2 text-center">
                    <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-nord14" onClick={() => attachToSession(session)}>Attach</span>
                    <span className="text-sm underline cursor-pointer inline-block mx-3 hover:text-nord11" onClick={() => killSession(session)}>Kill</span>
                </span>
            </div>)}
    </>;
});

function colorByStatus(status: SessionStatus): string {
    switch (status) {
        case SessionStatus.STARTED:
            return '#a3be8c';
        case SessionStatus.STARTING:
            return '#ebcb8b';
        case SessionStatus.STOPPING:
            return '#d08770';
        case SessionStatus.START_FAILED:
            return '#bf616a';
    }
}