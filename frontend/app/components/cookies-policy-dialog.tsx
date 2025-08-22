'use client'

import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  Cookie, 
  Shield, 
  Settings, 
  BarChart3, 
  Target, 
  Eye, 
  FileText,
  ExternalLink,
  ChevronDown,
  ChevronUp,
  Info,
  Mail,
  MapPin
} from 'lucide-react';
import envConfig from '../configs/envConfig';

interface CookiesPolicyDialogProps {
  children: React.ReactNode;
}

export const CookiesPolicyDialog: React.FC<CookiesPolicyDialogProps> = ({ children }) => {
  const [open, setOpen] = useState(false);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set());

  const toggleSection = (section: string) => {
    const newExpanded = new Set(expandedSections);
    if (newExpanded.has(section)) {
      newExpanded.delete(section);
    } else {
      newExpanded.add(section);
    }
    setExpandedSections(newExpanded);
  };

  const isExpanded = (section: string) => expandedSections.has(section);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-3 text-2xl">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Cookie className="h-6 w-6 text-primary" />
            </div>
            นโยบายการใช้คุกกี้ (Cookies Policy)
          </DialogTitle>
          <DialogDescription className="text-base">
            นโยบายการใช้งานคุกกี้ของปิยะวัฒน์ วงค์ญาติ เพื่อคุ้มครองข้อมูลส่วนบุคคลของคุณ
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Introduction */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">ข้อมูลทั่วไป</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-foreground leading-relaxed">
                นโยบายการใช้งานคุกกี้นี้ ("นโยบาย") อธิบายถึงวิธีการที่ปิยะวัฒน์ วงค์ญาติ 
                ("เรา", "ของเรา", "เรา") ใช้คุกกี้และเทคโนโลยีที่คล้ายคลึงกันบนเว็บไซต์ของเรา 
                เพื่อให้คุณเข้าใจถึงการใช้งานและสามารถควบคุมการตั้งค่าคุกกี้ได้
              </p>
              <p className="text-foreground leading-relaxed">
                การใช้เว็บไซต์ของเราแสดงว่าคุณยอมรับการใช้งานคุกกี้ตามที่อธิบายไว้ในนโยบายนี้
              </p>
            </CardContent>
          </Card>

          {/* What are Cookies */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <FileText className="h-5 w-5" />
                คุกกี้คืออะไร?
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-foreground leading-relaxed">
                คุกกี้เป็นไฟล์ข้อความขนาดเล็กที่ถูกเก็บไว้ในอุปกรณ์ของคุณเมื่อคุณเยี่ยมชมเว็บไซต์ 
                คุกกี้ช่วยให้เว็บไซต์จดจำข้อมูลเกี่ยวกับการเยี่ยมชมของคุณ เช่น ภาษาโปรด, 
                ขนาดตัวอักษร และการตั้งค่าอื่นๆ ที่สามารถช่วยปรับปรุงการใช้งานเว็บไซต์ของคุณ
              </p>
            </CardContent>
          </Card>

          {/* Cookie Categories */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Settings className="h-5 w-5" />
                ประเภทของคุกกี้ที่เราใช้
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Essential Cookies */}
              <div className="p-4 bg-muted/50 rounded-xl border border-border">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Shield className="h-5 w-5 text-primary" />
                    </div>
                    <div>
                      <h4 className="font-semibold text-foreground">คุกกี้ที่จำเป็น (Essential Cookies)</h4>
                      <p className="text-sm text-muted-foreground">จำเป็นสำหรับการทำงานพื้นฐานของเว็บไซต์</p>
                    </div>
                  </div>
                  <Badge variant="secondary">จำเป็น</Badge>
                </div>
                <div className="mt-3">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => toggleSection('essential')}
                    className="p-0 h-auto text-primary hover:text-primary/80 hover:bg-transparent"
                  >
                    {isExpanded('essential') ? (
                      <>
                        <ChevronUp className="h-4 w-4 mr-2" />
                        ซ่อนรายละเอียด
                      </>
                    ) : (
                      <>
                        <ChevronDown className="h-4 w-4 mr-2" />
                        แสดงรายละเอียด
                      </>
                    )}
                  </Button>
                  {isExpanded('essential') && (
                    <div className="mt-3 pt-3 border-t border-border">
                      <p className="text-sm text-foreground">
                        คุกกี้เหล่านี้จำเป็นสำหรับการทำงานของเว็บไซต์และไม่สามารถปิดการใช้งานได้ 
                        รวมถึงคุกกี้สำหรับการรักษาความปลอดภัย, การจัดการเซสชัน, และการทำงานพื้นฐานของเว็บไซต์
                      </p>
                    </div>
                  )}
                </div>
              </div>

              {/* Analytics Cookies */}
              <div className="p-4 bg-muted/50 rounded-xl border border-border">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-secondary/20 rounded-lg">
                      <BarChart3 className="h-5 w-5 text-secondary-foreground" />
                    </div>
                    <div>
                      <h4 className="font-semibold text-foreground">คุกกี้สำหรับการวิเคราะห์ (Analytics Cookies)</h4>
                      <p className="text-sm text-muted-foreground">ช่วยเราเข้าใจการใช้งานเว็บไซต์</p>
                    </div>
                  </div>
                  <Badge variant="outline">เลือกได้</Badge>
                </div>
                <div className="mt-3">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => toggleSection('analytics')}
                    className="p-0 h-auto text-primary hover:text-primary/80 hover:bg-transparent"
                  >
                    {isExpanded('analytics') ? (
                      <>
                        <ChevronUp className="h-4 w-4 mr-2" />
                        ซ่อนรายละเอียด
                      </>
                    ) : (
                      <>
                        <ChevronDown className="h-4 w-4 mr-2" />
                        แสดงรายละเอียด
                      </>
                    )}
                  </Button>
                  {isExpanded('analytics') && (
                    <div className="mt-3 pt-3 border-t border-border">
                      <p className="text-sm text-foreground">
                        คุกกี้เหล่านี้ช่วยเราเข้าใจว่าผู้เยี่ยมชมใช้เว็บไซต์อย่างไร 
                        รวมถึงจำนวนผู้เยี่ยมชม, หน้าที่ได้รับความนิยม, และการใช้งานทั่วไป 
                        ข้อมูลนี้ช่วยเราปรับปรุงเว็บไซต์ให้ดีขึ้น
                      </p>
                    </div>
                  )}
                </div>
              </div>

              {/* Marketing Cookies */}
              <div className="p-4 bg-muted/50 rounded-xl border border-border">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-accent/20 rounded-lg">
                      <Target className="h-5 w-5 text-accent-foreground" />
                    </div>
                    <div>
                      <h4 className="font-semibold text-foreground">คุกกี้สำหรับการตลาด (Marketing Cookies)</h4>
                      <p className="text-sm text-muted-foreground">สำหรับเนื้อหาที่เป็นส่วนตัวและโฆษณา</p>
                    </div>
                  </div>
                  <Badge variant="outline">เลือกได้</Badge>
                </div>
                <div className="mt-3">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => toggleSection('marketing')}
                    className="p-0 h-auto text-primary hover:text-primary/80 hover:bg-transparent"
                  >
                    {isExpanded('marketing') ? (
                      <>
                        <ChevronUp className="h-4 w-4 mr-2" />
                        ซ่อนรายละเอียด
                      </>
                    ) : (
                      <>
                        <ChevronDown className="h-4 w-4 mr-2" />
                        แสดงรายละเอียด
                      </>
                    )}
                  </Button>
                  {isExpanded('marketing') && (
                    <div className="mt-3 pt-3 border-t border-border">
                      <p className="text-sm text-foreground">
                        คุกกี้เหล่านี้ใช้เพื่อแสดงโฆษณาที่เกี่ยวข้องกับความสนใจของคุณ 
                        และเพื่อวัดประสิทธิภาพของแคมเปญโฆษณา 
                        ข้อมูลนี้อาจถูกแชร์กับพันธมิตรด้านโฆษณาของเรา
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* How to Control Cookies */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Eye className="h-5 w-5" />
                วิธีการควบคุมคุกกี้
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-foreground leading-relaxed">
                คุณสามารถควบคุมและจัดการคุกกี้ได้หลายวิธี:
              </p>
              <ul className="list-disc list-inside space-y-2 text-foreground">
                <li>ใช้การตั้งค่าคุกกี้ของเราในเว็บไซต์</li>
                <li>ปรับแต่งการตั้งค่าเบราว์เซอร์ของคุณ</li>
                <li>ใช้เครื่องมือจัดการคุกกี้ของเบราว์เซอร์</li>
                <li>ติดตั้งปลั๊กอินหรือส่วนขยายที่บล็อกคุกกี้</li>
              </ul>
              <div className="p-4 rounded-lg">
                <p className="text-sm text-black dark:text-white">
                  <strong>หมายเหตุ:</strong> การปิดการใช้งานคุกกี้อาจส่งผลต่อการทำงานของเว็บไซต์ 
                  และคุณอาจไม่สามารถใช้ฟีเจอร์บางอย่างได้
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Third Party Cookies */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">คุกกี้ของบุคคลที่สาม</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-foreground leading-relaxed">
                เว็บไซต์ของเราอาจใช้บริการจากบุคคลที่สาม เช่น Google Analytics, 
                Facebook Pixel หรือบริการโฆษณาอื่นๆ ซึ่งอาจใช้คุกกี้ของตนเอง 
                การใช้งานคุกกี้เหล่านี้อยู่ภายใต้นโยบายความเป็นส่วนตัวของบุคคลที่สามนั้นๆ
              </p>
            </CardContent>
          </Card>

          {/* Updates to Policy */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">การอัปเดตนโยบาย</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-foreground leading-relaxed">
                เราอาจอัปเดตนโยบายการใช้งานคุกกี้นี้เป็นครั้งคราว 
                การเปลี่ยนแปลงใดๆ จะถูกประกาศบนเว็บไซต์นี้ 
                เราขอแนะนำให้คุณตรวจสอบนโยบายนี้เป็นประจำ
              </p>
            </CardContent>
          </Card>

          {/* Contact Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">ข้อมูลติดต่อ</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-foreground leading-relaxed">
                หากคุณมีคำถามเกี่ยวกับนโยบายการใช้งานคุกกี้ของเรา 
                กรุณาติดต่อเราได้ที่:
              </p>
              <div className="mt-3 p-4 bg-muted/50 rounded-lg">
                <p className="font-medium text-foreground">{envConfig.contactPersonName}</p>
                <p className="text-sm text-muted-foreground">
                  อีเมล: {envConfig.email}<br />
                  ที่อยู่: กรุงเทพมหานคร, ประเทศไทย
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Footer */}
          <div className="pt-6 border-t border-border">
            <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
              <p className="text-sm text-muted-foreground text-center sm:text-left">
                อัปเดตล่าสุด: {new Date().toLocaleDateString('th-TH')}
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOpen(false)}
                >
                  ปิด
                </Button>
                <Button
                  size="sm"
                  onClick={() => {
                    setOpen(false);
                    // You can add navigation to cookie settings here
                  }}
                >
                  จัดการการตั้งค่าคุกกี้
                </Button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default CookiesPolicyDialog;
