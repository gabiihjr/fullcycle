import { Injectable } from "@nestjs/common";
import { PrismaService } from "src/prisma/services/prisma/prisma.service";

@Injectable()
export class AssetsService {
  constructor(private prismaService: PrismaService) {}

  create(data: { id: string; symbol: string; price: number }) {
    this.prismaService.asset.create({
      data,
    });
  }
}
