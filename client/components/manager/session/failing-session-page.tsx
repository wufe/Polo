import { IApp } from '@/state/models';
import { observer } from 'mobx-react-lite';
import React from 'react';
import { useParams } from 'react-router-dom';
import FailingSession from './failing-session';

type TProps = {
    app: IApp;
}
export const FailingSessionPage = observer((props: TProps) => {

    const params = useParams<{ uuid: string; }>();
    const uuid = params.uuid;

    return <FailingSession app={props.app} uuid={uuid} />
});

export default FailingSessionPage;