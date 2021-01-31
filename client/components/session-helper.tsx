import { APIRequestResult } from '@/api/common';
import { IApp, IStore } from '@/state/models';
import { observer } from 'mobx-react-lite';
import { Instance } from 'mobx-state-tree';
import React, { useEffect, useRef } from 'react';
import './session-helper.scss';
import { SessionHelperSession } from './session/session-helper-session';

type TProps = {
    store: IStore;
}

export const SessionHelperApp = observer((props: TProps) => {

    if (!props.store.app.session)
        return null;

    return <div className="session-helper__component">
        <SessionHelperSession session={props.store.app.session} />
    </div>
});