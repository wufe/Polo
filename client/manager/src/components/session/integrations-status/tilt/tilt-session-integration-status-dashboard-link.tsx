import React from 'react';
import { ISession } from "@polo/common/state/models";
import { observer } from "mobx-react-lite";

type TProps = {
    session: ISession;
    dashboards: ISession['integrations']['tilt']['dashboards'];
};

export const TiltSessionIntegrationStatusDashboardLink = observer((props: TProps) => {
    return <div className={'flex'}>
        {props.dashboards.map((dashboard, i) => {
            return <a
                key={dashboard.id}
                href={`${window.configuration.integrationsPublicURL}/tilt/${props.session.uuid}/${dashboard.id}`}
                target="_blank"
                className="text-xs uppercase px-3 py-2 rounded-md bg-nord14 text-nord5 hover:text-nord0-lighter hover:bg-nord14-alpha50 active:bg-nord14-darker active:text-nord5 transition-colors duration-200 ease-in-out mx-1">
                Tilt Dashboard {props.dashboards.length > 1 ? i + 1 : ''}
            </a>
        })}
    </div>;
});