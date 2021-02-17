import { APIRequestResult } from '@/api/common';
import { ISession, ISessionLog } from '@/state/models/session-model';
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
        mx-auto w-10/12 max-w-full flex flex-col min-w-0 min-h-0 flex-1 pt-3" style={{height:'calc(100vh - 120px)'}}>
        <div className="main-gradient-faded absolute left-0 right-0 top-0 pointer-events-none" style={{ bottom: `${overlayBottom}%`}}></div>
        <h1 className="text-4xl mb-3 font-quicksand font-light text-nord1 dark:text-nord5 z-10">Session</h1>
        <div className="text-lg text-gray-500 mb-7 font-quicksand z-10">Id: {props.session.uuid}</div>
        <SessionLogs
            logs={values(props.session.logs) as any as ISessionLog[]}
            onLogsProportionChanged={setOverlayProportions} />
    </div>
});

export default Session;