import { ISession } from '@/state/models';
import { observer } from 'mobx-react-lite';
import React from 'react';
import { ApplicationSession } from '../session/application-session';
import './application-sessions.scss';

type TProps = {
    sessions: ISession[];
}

export const ApplicationSessions = observer((props: TProps) => {
    const visibleSessions = props.sessions
        .filter(session => !session.replacesSession);

    return <div className="flex flex-col items-stretch">
        {/* <h4 className="my-1 text-xs text-gray-500 uppercase">Sessions:</h4> */}
        <input type="text" name="" id="" placeholder="Filter sessions" className="bg-transparent border border-gray-200 dark:border-gray-500 text-sm py-2 px-3 rounded-md mb-1" />
        {visibleSessions.length && <span className="text-xs lg:text-sm text-gray-500 pl-2">{visibleSessions.length} sessions</span>}
        <div className="mt-3">
            {visibleSessions
                .map((session, key) => <ApplicationSession session={session} key={key} />)}
        </div>
        <div className="mt-7 mb-0 flex justify-center">
            <div className="min-w-9/12 h-1 border-b border-gray-300 dark:border-gray-500"></div>
        </div>
    </div>;
});