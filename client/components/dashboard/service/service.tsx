import { IService } from '@/state/models/service';
import React from 'react';
import './service.scss';

type TProps = {
    service: IService;
}

export const Service = (props: TProps) => <div className="service__component">
    <h3>{props.service.name}</h3>
    <div>{props.service.remote}</div>
    <div>{props.service.target}</div>
</div>