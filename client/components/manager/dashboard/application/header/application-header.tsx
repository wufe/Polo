import { useModal } from '@/components/manager/modal/modal-hooks';
import { IApplication, ISession } from '@/state/models';
import { observer } from 'mobx-react-lite';
import { values } from 'mobx';
import React from 'react';
import { ApplicationOptionsModal } from '../options/application-options-modal';
import './application-header.scss';
import { useHistory } from 'react-router';
import dayjs from 'dayjs';
import { Button } from '@/components/shared/elements/button/button';
import { MenuIcon } from '@/components/shared/elements/icons/menu/menu-icon';

type TProps = {
    name: string;
    filename: string;

    failedSessions: ISession[] | null;
}
export const ApplicationHeader = (props: TProps) => {
    const { show, hide } = useModal();
    const history = useHistory();

    const anyFailedSession = props.failedSessions && props.failedSessions.length > 0;

    const applicationOptionsModalName = `application-options-${props.name}`;

    const goTo = (session: ISession) => {
        hide();
        history.push(`/_polo_/session/failing/${session.uuid}`);
    }

    return <div className="application-header">
        <div className="flex justify-between min-w-0 max-w-full flex-nowrap items-center">
            <h3 className="text-xl lg:text-2xl leading-5 font-bold overflow-hidden overflow-ellipsis whitespace-nowrap flex-grow flex-shrink pr-6" title={props.name}>{props.name}</h3>
            <Button
                ghost
                largeIcon
                onClick={() => show(applicationOptionsModalName)}
                icon={<MenuIcon />}>
                {anyFailedSession && <div className="__error-circle"></div>}
            </Button>
        </div>
        <span className="text-gray-400 text-sm">{props.filename}</span>

        <ApplicationOptionsModal
            modalName={applicationOptionsModalName}
            applicationName={props.name}
            failedSessions={sortSessionsByCreationTimeDesc(props.failedSessions)}
            onSessionClick={goTo} />
    </div>
};

function sortSessionsByCreationTimeDesc(sessions: ISession[] | null): ISession[] {
    if (!sessions) return sessions;
    return sessions
        .sort((a, b) => {
            const dateA = dayjs(a.createdAt);
            const dateB = dayjs(b.createdAt);
            if (dateA.isBefore(dateB))
                return 1;
            if (dateA.isAfter(dateB))
                return -1;
            return 0;
        })
}