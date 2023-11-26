import { useModal } from '@/components/modal/modal-hooks';
import { IApplication, ISession } from '@polo/common/state/models';
import { observer } from 'mobx-react-lite';
import { values } from 'mobx';
import React from 'react';
import { ApplicationOptionsModal } from '../options/application-options-modal';
import './application-header.scss';
import { useHistory } from 'react-router';
import dayjs from 'dayjs';
import { Button } from '@polo/common/components/elements/button/button';
import { MenuIcon } from '@polo/common/components/elements/icons/menu/menu-icon';
import { FailureStatus, TFailuresDictionary } from '@polo/common/state/models/failures-model';

type TProps = {
    name: string;
    filename: string;

    showApplicationSelector?: boolean;
    onApplicationsSelectorClick?: () => void;

    failures: TFailuresDictionary | null;
}
export const ApplicationHeader = (props: TProps) => {
    const { show, hide } = useModal();
    const history = useHistory();

    const anyUnacknowledged = props.failures && props.failures.unacknowledged.length > 0;

    const applicationOptionsModalName = `application-options-${props.name}`;

    const goTo = (session: ISession) => {
        hide();
        history.push(`/_polo_/session/failing/${session.uuid}`);
    }

    const onApplicationHeaderClick = () => {
        if (props.showApplicationSelector) {
            props.onApplicationsSelectorClick?.();
        }
    };

    const onApplicationOptionsClick = (event: React.MouseEvent) => {
        event.stopPropagation();
        show(applicationOptionsModalName);
    };

    return <div className="application-header">
        <div className="flex justify-between min-w-0 max-w-full flex-nowrap">
            <div className="flex-grow flex-shrink pr-6 min-w-0" onClick={onApplicationHeaderClick}>
                <h3 className="text-xl lg:text-2xl leading-10 font-bold overflow-hidden overflow-ellipsis whitespace-nowrap __title" title={props.name}>{props.name}</h3>
                <span className="text-gray-400 text-sm __filename">{props.filename}</span>
            </div>
            <Button
                ghost
                largeIcon
                onClick={onApplicationOptionsClick}
                icon={<MenuIcon />}>
                {anyUnacknowledged && <div className="__error-circle"></div>}
            </Button>
        </div>


        <ApplicationOptionsModal
            modalName={applicationOptionsModalName}
            applicationName={props.name}
            failures={props.failures}
            onSessionClick={goTo} />
    </div>
};

