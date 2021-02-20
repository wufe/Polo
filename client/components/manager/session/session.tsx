import { APIRequestResult } from '@/api/common';
import { ISession, ISessionLog } from '@/state/models/session-model';
import dayjs from 'dayjs';
import { values } from 'mobx';
import { observer } from 'mobx-react-lite';
import React, { useEffect, useRef, useState } from 'react';
import { useHistory } from 'react-router-dom';
import { SessionLogs } from './session-logs';
import { useSessionRetrieval } from './session-retrieval-hook';

type TProps = {
    session: ISession;
}
export const Session = observer((props: TProps) => {

    const [overlayBottom, setOverlayBottom] = useState(100);
    
    useSessionRetrieval(props.session);

    const setOverlayProportions = (proportions: number) => {
        const percentage = parseInt(`${proportions * 100}`);
        const inversePercentage = 100 - percentage;
        setOverlayBottom(inversePercentage);
    }

    return <div className="
        mx-auto w-10/12 max-w-full flex flex-col min-w-0 min-h-0 flex-1 pt-3 font-quicksand" style={{height:'calc(100vh - 120px)'}}>
        <div className="main-gradient-faded absolute left-0 right-0 top-0 pointer-events-none" style={{ bottom: `${overlayBottom}%`}}></div>
        <h1 className="text-4xl mb-3 font-quicksand font-light text-nord1 dark:text-nord5 z-10">Session</h1>
        <div className="text-lg text-nord1 dark:text-nord5 mb-4 z-10 border-l pl-3 border-gray-500">
            <span>{props.session.checkout}</span>
        </div>
        <blockquote className="relative px-4 py-3 italic text-gray-500 dark:text-gray-400 z-10 bg-nord6 shadow-md dark:bg-nord-5 leading-loose">
            <p className="text-sm pb-1">{props.session.commitMessage}</p>
            <cite className="flex items-center">
                <span className="mb-1 text-sm font-bold italic flex-1">~ {props.session.commitAuthorName}</span>
                <span className="mb-1 text-xs font-light italic">({props.session.commitAuthorEmail} - {dayjs(props.session.commitDate).format('DD MMM HH:mm')})</span>
            </cite>
        </blockquote>
        <SessionLogs
            logs={values(props.session.logs) as any as ISessionLog[]}
            onLogsProportionChanged={setOverlayProportions} />
    </div>
});

export default Session;