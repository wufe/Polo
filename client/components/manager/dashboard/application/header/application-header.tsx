import { useModal } from '@/components/manager/modal/modal-hooks';
import { IApplication, ISession } from '@/state/models';
import { observer } from 'mobx-react-lite';
import { values } from 'mobx';
import React from 'react';
import { ApplicationOptionsModal } from '../options/application-options-modal';
import './application-header.scss';
import { useHistory } from 'react-router';
import dayjs from 'dayjs';
import { Button } from '@/components/shared/ui-elements/button/button';
import { MenuIcon } from '@/components/shared/ui-elements/icons/menu/menu-icon';
import { FailureStatus, TFailuresDictionary } from '@/state/models/failures-model';

type TProps = {
    id: string;
    name: string;
    filename: string
    failures: TFailuresDictionary | null;
}
export const ApplicationHeader = (props: TProps) => {
    const { show, hide } = useModal();
    const history = useHistory();

    const anyUnacknowledged = props.failures && props.failures.unacknowledged.length > 0;

    const applicationOptionsModalName = `application-options-${props.name}`;

    const goToFailingSession = (session: ISession) => {
        hide();
        history.push(`/_polo_/session/failing/${session.uuid}`);
    }

    const goToApplicationConfigurationEditPage = () => {
        hide();
        history.push(`/_polo_/application/${props.id}/edit`);
    }

    return <div className="application-header">
        <div className="flex justify-between min-w-0 max-w-full flex-nowrap items-center">
            <h3 className="text-xl lg:text-2xl leading-5 font-bold overflow-hidden overflow-ellipsis whitespace-nowrap flex-grow flex-shrink pr-6" title={props.name}>{props.name}</h3>
            <Button
                ghost
                largeIcon
                onClick={() => show(applicationOptionsModalName)}
                icon={<MenuIcon />}>
                {anyUnacknowledged && <div className="__error-circle"></div>}
            </Button>
        </div>
        <span className="text-gray-400 text-sm">{props.filename}</span>

        <ApplicationOptionsModal
            modalName={applicationOptionsModalName}
            applicationName={props.name}
            failures={props.failures}
            onFailingSessionClick={goToFailingSession}
            onApplicationConfigurationEditClick={goToApplicationConfigurationEditPage} />
    </div>
};

